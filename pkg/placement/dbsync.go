/*

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

package placement

import (
	placementv1 "github.com/openstack-k8s-operators/placement-operator/api/v1beta1"

	common "github.com/openstack-k8s-operators/lib-common/modules/common"
	env "github.com/openstack-k8s-operators/lib-common/modules/common/env"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// DbSyncJob func
func DbSyncJob(
	instance *placementv1.PlacementAPI,
	labels map[string]string,
	annotations map[string]string,
) *batchv1.Job {
	args := []string{"-c"}
	if instance.Spec.Debug.DBSync {
		args = append(args, common.DebugCommand)
	} else {
		args = append(args, KollaServiceCommand)
	}

	envVars := map[string]env.Setter{}
	envVars["KOLLA_CONFIG_STRATEGY"] = env.SetValue("COPY_ALWAYS")
	envVars["KOLLA_BOOTSTRAP"] = env.SetValue("true")

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName + "-db-sync",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ServiceAccountName: instance.RbacResourceName(),
					Containers: []corev1.Container{
						{
							Name: ServiceName + "-db-sync",
							Command: []string{
								"/bin/bash",
							},
							Args:  args,
							Image: instance.Spec.ContainerImage,
							SecurityContext: &corev1.SecurityContext{
								RunAsUser: ptr.To(PlacementUserID),
							},
							Env:          env.MergeEnvs([]corev1.EnvVar{}, envVars),
							VolumeMounts: getVolumeMounts("dbsync"),
						},
					},
				},
			},
		},
	}

	job.Spec.Template.Spec.Volumes = getVolumes(ServiceName)

	initContainerDetails := APIDetails{
		ContainerImage:       instance.Spec.ContainerImage,
		DatabaseHost:         instance.Status.DatabaseHostname,
		DatabaseUser:         instance.Spec.DatabaseUser,
		DatabaseName:         DatabaseName,
		OSPSecret:            instance.Spec.Secret,
		DBPasswordSelector:   instance.Spec.PasswordSelectors.Database,
		UserPasswordSelector: instance.Spec.PasswordSelectors.Service,
		VolumeMounts:         getInitVolumeMounts(),
	}
	job.Spec.Template.Spec.InitContainers = initContainer(initContainerDetails)

	return job
}
