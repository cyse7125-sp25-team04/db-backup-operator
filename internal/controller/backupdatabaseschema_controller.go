package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	backupv1alpha1 "test.io/test-crd/api/v1alpha1"
)

// BackupDatabaseSchemaReconciler reconciles a BackupDatabaseSchema object
type BackupDatabaseSchemaReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="",resources=events,verbs=create;update;patch
// +kubebuilder:rbac:groups=backup.test.io,resources=backupdatabaseschemas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=backup.test.io,resources=backupdatabaseschemas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=backup.test.io,resources=backupdatabaseschemas/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

func (r *BackupDatabaseSchemaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("BackupDatabaseSchema", req.NamespacedName)
	log.Info("Starting reconciliation")

	// 1. Fetch the BackupDatabaseSchema instance
	var backupCR backupv1alpha1.BackupDatabaseSchema
	if err := r.Get(ctx, req.NamespacedName, &backupCR); err != nil {
		if errors.IsNotFound(err) {
			log.Info("BackupDatabaseSchema resource not found; it might have been deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch BackupDatabaseSchema")
		return ctrl.Result{}, err
	}
	log.Info("Fetched BackupDatabaseSchema", "spec", backupCR.Spec)

	// 2. Check if a CronJob already exists in the desired namespace (using a label or name)
	cronJobName := fmt.Sprintf("backup-%s-cronjob", backupCR.Name)
	cronJobNamespace := backupCR.Spec.BackupJobNamespace
	var existingCronJob batchv1.CronJob
	err := r.Get(ctx, client.ObjectKey{Namespace: cronJobNamespace, Name: cronJobName}, &existingCronJob)
	if err == nil {
		// CronJob already exists, so we do nothing (or update it if needed)
		log.Info("CronJob already exists; skipping creation", "cronJobName", cronJobName)
		return ctrl.Result{}, nil
	} else if !errors.IsNotFound(err) {
		log.Error(err, "Error checking for existing CronJob")
		return ctrl.Result{}, err
	}

	// 3. Pre-generate a timestamp and backup file name (for status update)
	// backupFileName := fmt.Sprintf("%s_%d.sql", backupCR.Spec.DBSchema, time.Now().UnixNano())

	// 4. Build the CronJob object with schedule "*/5 * * * *" (every 5 minutes)
	cmd := `backupFile=/backup/${DB_SCHEMA}_$(date +\%s%N).sql && ` +
		`echo "Dumping schema from ${DB_HOST}:${DB_PORT}" && ` +
		`mysqldump --host=${DB_HOST} --port=${DB_PORT} --user=${DB_USER} --password=${DB_PASSWORD} --no-data ${dbName} > $backupFile && ` +
		`echo "Backup created at $backupFile" && ` +
		`gcloud config set project gcp-dev-7125 && ` +
		`gcloud auth activate-service-account --key-file=/var/secrets/gcp/key.json && ` +
		// `sleep 50000 &&` +
		`gsutil cp $backupFile gs://${GCS_BUCKET}/$backupFile`

	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cronJobName,
			Namespace: cronJobNamespace,
			Labels: map[string]string{
				"backup-cr": backupCR.Name,
			},
		},
		Spec: batchv1.CronJobSpec{
			// Set the schedule to every 5 minutes
			Schedule: "*/5 * * * *",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							ImagePullSecrets: []corev1.LocalObjectReference{
								{
									Name: "myapp-secret-docker",
								},
							},
							Containers: []corev1.Container{
								{
									Name:            "backup",
									Image:           "csye712504/db-backup-operator:latest",
									Command:         []string{"/bin/sh", "-c"},
									Args:            []string{cmd},
									Env: []corev1.EnvVar{
										{Name: "DB_SCHEMA", Value: backupCR.Spec.DBSchema},
										{Name: "GCS_BUCKET", Value: backupCR.Spec.GCSBucket},
										{Name: "DB_HOST", Value: backupCR.Spec.DBHost},
										{Name: "DB_PORT", Value: fmt.Sprintf("%d", backupCR.Spec.DBPort)},
										{Name: "DB_USER", Value: backupCR.Spec.DBUser},
										{Name: "DB_NAME", Value: backupCR.Spec.DBName},
										// Uncomment if you want to pull DB_PASSWORD from a secret.
										{
											Name: "DB_PASSWORD",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: backupCR.Spec.DBPasswordSecretName,
													},
													Key: backupCR.Spec.DBPasswordSecretKey,
												},
											},
										},
										{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "/var/secrets/gcp/key.json"},
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "backup-volume",
											MountPath: "/backup",
										},
										{
											Name:      "gcp-key",
											MountPath: "/var/secrets/gcp",
											ReadOnly:  true,
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "backup-volume",
									VolumeSource: corev1.VolumeSource{
										EmptyDir: &corev1.EmptyDirVolumeSource{},
									},
								},
								{
									Name: "gcp-key",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: backupCR.Spec.GCPServiceAccount, // e.g., "gcp-key"
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// 5. Create the CronJob
	log.Info("Creating new backup CronJob", "cronJob", cronJob)
	if err := r.Create(ctx, cronJob); err != nil {
		log.Error(err, "Failed to create backup CronJob")
		return ctrl.Result{}, err
	}
	log.Info("Backup CronJob created successfully", "cronJobName", cronJobName)

	// 6. Update the CRD status with backup details.
	backupCR.Status.RecentJobName = cronJobName
	backupCR.Status.LastBackupTime = time.Now().UTC().Format(time.RFC3339)
	backupCR.Status.JobStatus = "CronJobCreated"
	backupCR.Status.BackupLocation = fmt.Sprintf("gs://%s/%s_$(date +%%s%%N).sql", backupCR.Spec.GCSBucket, backupCR.Spec.DBSchema)
	if err := r.Status().Update(ctx, &backupCR); err != nil {
		log.Error(err, "Failed to update BackupDatabaseSchema status")
		return ctrl.Result{}, err
	}
	log.Info("Updated BackupDatabaseSchema status", "status", backupCR.Status)

	// 7. No requeue since the CronJob is created and will manage scheduling.
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BackupDatabaseSchemaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log = ctrl.Log.WithName("controllers").WithName("BackupDatabaseSchema")
	return ctrl.NewControllerManagedBy(mgr).
		For(&backupv1alpha1.BackupDatabaseSchema{}).
		Complete(r)
}
