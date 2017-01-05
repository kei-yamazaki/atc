package tests_test

import (
	"fmt"
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/atc/cessna"
	. "github.com/concourse/atc/cessna/resource"
	"github.com/concourse/baggageclaim"

	bclient "github.com/concourse/baggageclaim/client"

	"testing"
)

var (
	testBaseResource Resource
	testWorker       cessna.Worker
	baseResourceType ResourceType
	workerIp         = "10.244.16.2"
	tarPath          = "./echo-resource-rootfs.tar"
)

var _ = BeforeSuite(func() {
	if os.Getenv("WORKER_IP") != "" {
		workerIp = os.Getenv("WORKER_IP")
	}

	if os.Getenv("TAR_PATH") != "" {
		tarPath = os.Getenv("TAR_PATH")
	}

	Expect(workerIp).ToNot(BeEmpty())
	Expect(tarPath).ToNot(BeEmpty())

	testWorker = cessna.NewWorker(fmt.Sprintf("%s:7777", workerIp), fmt.Sprintf("http://%s:7788", workerIp))

	rootFSPath, err := createBaseResourceVolume(tarPath)

	Expect(err).ToNot(HaveOccurred())

	baseResourceType = ResourceType{
		RootFSPath: rootFSPath,
		Name:       "echo",
	}

})

func TestResource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Resource Suite")
}

func createBaseResourceVolume(tarPath string) (string, error) {
	baggageclaimClient := bclient.New(fmt.Sprintf("http://%s:7788", workerIp), http.DefaultTransport)

	volumeSpec := baggageclaim.VolumeSpec{
		Strategy:   baggageclaim.EmptyStrategy{},
		Privileged: true,
	}

	volume, err := baggageclaimClient.CreateVolume(
		lager.NewLogger("create-volume-for-base-resource"),
		"base-resource-volume",
		volumeSpec)

	if err != nil {
		return "", err
	}

	tarfile, err := os.Open(tarPath)
	if err != nil {
		return "", err
	}

	err = volume.StreamIn("/", tarfile)
	if err != nil {
		return "", err
	}

	return volume.Path(), nil
}
