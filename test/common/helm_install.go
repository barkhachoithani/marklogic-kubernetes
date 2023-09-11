package common

import (
	"crypto/tls"
	"path/filepath"
	"fmt"
	"testing"
	"time"
	"strings"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/random"
)

var NamespaceName = "marklogic-" + strings.ToLower(random.UniqueId())
var KubectlOpt = k8s.NewKubectlOptions("", "", NamespaceName)

func HelmInstallFn(t *testing.T, options map[string]string, releaseName string) string {
	// Path to the helm chart we will test
	helmChartPath, e := filepath.Abs("../../charts")
	if e != nil {
		t.Fatalf(e.Error())
	}
	t.Logf("====Creating namespace: " + NamespaceName)
	k8s.CreateNamespace(t, KubectlOpt, NamespaceName)

	helmOptions := &helm.Options{
		KubectlOptions: KubectlOpt,
		SetValues: options,
	}
	t.Logf("====Installing Helm Chart")
	helm.Install(t, helmOptions, helmChartPath, releaseName)

	tlsConfig := tls.Config{}
	podName := releaseName + "-marklogic-0"
	// wait until the pod is in Ready status
	k8s.WaitUntilPodAvailable(t, KubectlOpt, podName, 10, 15*time.Second)
	tunnel7997 := k8s.NewTunnel(KubectlOpt, k8s.ResourceTypePod, podName, 7997, 7997)
	defer tunnel7997.Close()
	tunnel7997.ForwardPort(t)
	endpoint7997 := fmt.Sprintf("http://%s", tunnel7997.Endpoint())

	// verify if 7997 health check endpoint returns 200
	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		endpoint7997,
		&tlsConfig,
		10,
		15*time.Second,
		func(statusCode int, body string) bool {
			return statusCode == 200
		},
	)
	return podName
}
