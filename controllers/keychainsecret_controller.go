/*
Copyright (c) 2020 Facebook, Inc. and its affiliates.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"os"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	aqueductv1 "github.com/davidewatson/keychain/api/v1"
	"github.com/davidewatson/keychain/pkg/command"
)

const ()

// KeychainSecretReconciler reconciles a KeychainSecret object
type KeychainSecretReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=aqueduct.k8s.facebook.com,resources=keychainsecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aqueduct.k8s.facebook.com,resources=keychainsecrets/status,verbs=get;update;patch

// Reconcile is called when a watched resource needs to be reconciled.
func (r *KeychainSecretReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var keychainSecret aqueductv1.KeychainSecret

	ctx := context.Background()
	log := r.Log.WithValues("keychainsecret", req.NamespacedName)

	// Get current version of the spec.
	if err := r.Get(ctx, req.NamespacedName, &keychainSecret); err != nil {
		log.Error(err, "unable to fetch KeychainSecret")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	identity, err := r.GetOrCreateIdentity(ctx, keychainSecret)
	if err != nil {
		panic("TODO: Implement core admission controller to fail on Namespace creation instead of panicking here...")
	}

	_, err = r.CreateSecretFromKeychain(ctx, identity, keychainSecret)
	if err != nil {
		return ctrl.Result{RequeueAfter: 5 * time.Minute}, err
	}

	duration, err := time.ParseDuration(keychainSecret.Spec.TTL)
	if err != nil {
		panic("TTL was not a valid Duration. This should have been caught during validation!?")
	}
	return ctrl.Result{RequeueAfter: duration}, nil
}

// CreateSecretFromKeychain creates a Kubernetes Secret corresponding to the KeychainSecret.
// It will also update an existing Secret if the TTL has expired, for rotation purposes.
func (r *KeychainSecretReconciler) CreateSecretFromKeychain(ctx context.Context, identitySecret *corev1.Secret, keychainSecret aqueductv1.KeychainSecret) (*corev1.Secret, error) {
	newSecret := &corev1.Secret{}

	log := r.Log.WithValues("CreateSecretFromKeychain") //, req.NamespacedName)

	// Get current Secret, if any.
	originalSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: keychainSecret.ObjectMeta.Namespace, Name: keychainSecret.Spec.Name}, originalSecret); err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "unable to fetch Secret")
			return nil, err
		}
	}

	// Get the keychain secret
	secret, err := command.GetKeychainSecret(ctx, command.GetKeychainSecretParams{
		Group: keychainSecret.Spec.Group,
		Name:  keychainSecret.Spec.Name,
	})
	if err != nil {
		return nil, err
	}
	data := map[string][]byte{keychainSecret.Spec.Name: secret}

	duration, _ := time.ParseDuration(keychainSecret.Spec.TTL)

	// Either we need to create the secret, or we need to refresh it.
	if originalSecret == nil {
		newSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: keychainSecret.ObjectMeta.Namespace,
				Name:      keychainSecret.Spec.Name,
			},
			Data: data,
		}
		err := r.Create(ctx, newSecret)
		if err != nil {
			return nil, err
		}
	} else if keychainSecret.Status.LastUpdate.Add(duration).After(time.Now()) {
		newSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: keychainSecret.ObjectMeta.Namespace,
				Name:      keychainSecret.Spec.Name,
			},
			Data: data,
		}
		err := r.Patch(ctx, newSecret, client.MergeFrom(originalSecret))
		if err != nil {
			return nil, err
		}
	}

	return newSecret, nil
}

// GetOrCreateIdentity gets or creates a certificate and stores it in a Secret within the controllers namespace.
// NOTE: This is a hack (maybe) until we have an admission webhook. Since Namespaces are core types and not CRDs,
// kubebuilder will be of little help in creating such a webhook. See here for more: https://github.com/kubernetes-sigs/controller-runtime/tree/master/examples
// TODO: Document the pros and cons of using an admission webhook.
func (r *KeychainSecretReconciler) GetOrCreateIdentity(ctx context.Context, keychainSecret aqueductv1.KeychainSecret) (*corev1.Secret, error) {
	// We store the identity within the controllers namespace and not the KeychainSecrets namespace. This is so we may
	// control who may create Secrets from KeychainSecrets.
	controllerNamespace := os.Getenv("CONTROLLER_NAMESPACE")
	// Namespace names are unique within a cluster, so this name will be unique as well...
	certName := keychainSecret.ObjectMeta.Namespace

	// Check if we already have an identity
	identitySecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: controllerNamespace, Name: certName}, identitySecret); err != nil {
		// Either we don't have an identity, or there was another error.
		if client.IgnoreNotFound(err) != nil {
			// Something else is wrong, give up!
			return nil, err
		}

		// We need to create an identity
		cert, err := command.ProvisionServiceIdentity(ctx, command.ProvisionServiceIdentityParams{
			Algorithm: "rsa:4096",
			Days:      365,
			Subject:   "'/CN=judkins.house/O=Facebook/C=US'",
		})
		if err != nil {
			return nil, err
		}

		data := map[string][]byte{"cert": cert}

		identitySecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: controllerNamespace,
				Name:      certName,
			},
			Data: data,
		}
		err = r.Create(ctx, identitySecret)
		if err != nil {
			return nil, err
		}
	}

	return identitySecret, nil
}

// SetupWithManager sets up the controller with manager.
func (r *KeychainSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aqueductv1.KeychainSecret{}).
		Complete(r)
}
