package main

import (
	"context"
	"flag"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"strings"
)

var (
	application = flag.String("a", "", "Application name")
	version     = flag.String("v", "latest", "Application version")
	env         = flag.String("e", "dev", "Deployment env, default dev")
)

var config *rest.Config
var client *kubernetes.Clientset
var err error

const REPO string = "nexus3.showcai.com.cn:5000"

var NAMESPACE = "dev"
var APOLLO = "http://10.141.48.10:18080/"

func main() {
	flag.Parse()

	home := homedir.HomeDir()
	config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
	if err != nil {
		panic(err)
	}

	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	if *application == "" {
		panic("application is empty")
	}

	if *env != "" {
		NAMESPACE = *env
	}

	if *env == "sit" {
		APOLLO = "http://10.141.48.10:28080/"
	}

	println("application=" + *application + ", version=" + *version + ", env=" + *env)
	deployment(*application, *version)
	service(*application)
}

func deployment(application string, version string) {
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      application,
			Namespace: NAMESPACE,
			Labels: map[string]string{
				"app":                       application,
				"app.kubernetes.io/name":    application,
				"app.kubernetes.io/version": "v1",
			},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":                       application,
					"app.kubernetes.io/name":    application,
					"app.kubernetes.io/version": "v1",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                       application,
						"app.kubernetes.io/name":    application,
						"app.kubernetes.io/version": "v1",
					},
				},
				Spec: v1.PodSpec{
					ImagePullSecrets: []v1.LocalObjectReference{
						{Name: "dw-secret"},
					},
					Containers: []v1.Container{
						{
							Name:  application,
							Image: REPO + "/" + application + ":" + version,
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "APP_ID",
									Value: application,
								},
								{
									Name:  "ENV",
									Value: strings.ToUpper(NAMESPACE),
								},
								{
									Name:  "APOLLO_CONFIGSERVICE",
									Value: APOLLO,
								},
								{
									Name:  "SW_AGENT_COLLECTOR_BACKEND_SERVICES",
									Value: "10.141.48.10:11800",
								},
								{
									Name:  "SW_AGENT_NAME",
									Value: application,
								},
							},
							ImagePullPolicy: v1.PullAlways,
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									"cpu":    resource.MustParse("2000m"),
									"memory": resource.MustParse("2Gi"),
								},
								Requests: v1.ResourceList{
									"cpu":    resource.MustParse("1000m"),
									"memory": resource.MustParse("500Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	err = client.AppsV1().Deployments(NAMESPACE).Delete(context.TODO(), application, metav1.DeleteOptions{})
	if err != nil {

	}
	_, err = client.AppsV1().Deployments(NAMESPACE).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

func int32Ptr(i int32) *int32 { return &i }

func service(application string) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      application,
			Namespace: NAMESPACE,
			Labels: map[string]string{
				"app":                       application,
				"app.kubernetes.io/name":    application,
				"app.kubernetes.io/version": "v1",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:     "http-8080",
					Protocol: v1.ProtocolTCP,
					Port:     8080,
					TargetPort: intstr.IntOrString{
						IntVal: 8080,
					},
				},
			},
			Selector: map[string]string{
				"app":                       application,
				"app.kubernetes.io/name":    application,
				"app.kubernetes.io/version": "v1",
			},
		},
	}
	_, err = client.CoreV1().Services(NAMESPACE).Create(context.TODO(), svc, metav1.CreateOptions{})
	if err != nil {

	}
}
