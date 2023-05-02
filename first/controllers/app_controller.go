/*
Copyright 2023.

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
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	demov1 "github.com/candy44777/operator-app/firest/api/v1"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=demo.candy-box.top,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=demo.candy-box.top,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=demo.candy-box.top,resources=apps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the App object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	app := &demov1.App{}
	// 在使用 get 获取对象的时候，会把缓存中的值赋值给 app，如果缓存中没有，会从 etcd 中获取，并且会把获取到的值放入缓存中
	// 这样下次获取的时候就可以直接从缓存中获取了，不需要再从 etcd 中获取，这样就提高了效率，但是会有数据不一致的问题
	// 也就是说，如果你在 etcd 中修改了对象的值，但是缓存中的值还是旧的，这样就会导致数据不一致的问题，所以在使用 get 的时候
	// 一定要注意这个问题，如果你的业务逻辑不允许数据不一致，那么就不要使用 get，而是使用 list，这样就不会有数据不一致的问题
	// 但是使用 list 会导致效率降低，因为 list 每次都会从 etcd 中获取，不会从缓存中获取，所以在使用的时候要根据实际情况来选择
	if err := r.Client.Get(ctx, req.NamespacedName, app); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// 由于在修改 app 的时候，会把缓存中的值赋值给 app，所以在修改的时候，一定要使用 DeepCopy 方法，这样就不会修改缓存中的值了
	// 也就不会导致数据不一致的问题了，在修改完成后，需要调用 Update 方法，把修改后的值更新到 etcd 中
	appCopy := app.DeepCopy()

	action := appCopy.Spec.Action
	object := appCopy.Spec.Object

	ret := strings.Join([]string{action, ",", object}, "")
	appCopy.Status.Result = ret

	// 这里只是更新了 status 中的值，app 并没有要更新的数据，使用该方法会触发一次 event
	// 会在调用一次 reconcile 方法，所以在修改的时候，一定要注意，不要修改不需要修改的值，否则会导致死循环
	// r.Client.Update(ctx, appCopy, &client.UpdateOptions{})

	// 只更新 status 中的值，不会触发 event，也不会调用 reconcile 方法
	if err := r.Status().Update(ctx, appCopy, &client.SubResourceUpdateOptions{}); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demov1.App{}).
		Complete(r)
}
