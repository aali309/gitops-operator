package secret

import (
	"bytes"
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	matcher "github.com/onsi/gomega/types"
	"github.com/redhat-developer/gitops-operator/test/openshift/e2e/ginkgo/fixture/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Update will keep trying to update object until it succeeds, or times out.
func Update(obj *corev1.Secret, modify func(*corev1.Secret)) {
	k8sClient, _ := utils.GetE2ETestKubeClient()

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of the object
		err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(obj), obj)
		if err != nil {
			return err
		}

		modify(obj)

		// Attempt to update the object
		return k8sClient.Update(context.Background(), obj)
	})
	Expect(err).ToNot(HaveOccurred())
}

// Update will keep trying to update object until it succeeds, or times out.
func UpdateWithError(obj *corev1.Secret, modify func(*corev1.Secret)) error {
	k8sClient, _ := utils.GetE2ETestKubeClient()

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of the object
		err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(obj), obj)
		if err != nil {
			return err
		}

		modify(obj)

		// Attempt to update the object
		return k8sClient.Update(context.Background(), obj)
	})

	return err
}

// HaveStringDataKeyValue returns true if Secret has 'key' field under .data map, and the value of that field is equal to 'value'
func HaveStringDataKeyValue(key string, value string) matcher.GomegaMatcher {
	return fetchSecret(func(sec *corev1.Secret) bool {
		a, exists := sec.Data[key]
		if !exists {
			GinkgoWriter.Println("HaveStringDataKeyValue - Key:", key, "does not exist")
			return false
		}

		GinkgoWriter.Println("HaveStringDataKeyValue - Key:", key, "Expected:", value, "/ Have:", string(a))

		return string(a) == value
	})

}

// HaveDataKeyValue returns true if Secret has 'key' field under .data map, and the value of that field is equal to 'value'
func HaveDataKeyValue(key string, value []byte) matcher.GomegaMatcher {
	return fetchSecret(func(sec *corev1.Secret) bool {
		a, exists := sec.Data[key]
		if !exists {
			return false
		}
		return bytes.Equal(a, value)
	})

}

// This is intentionally NOT exported, for now. Create another function in this file/package that calls this function, and export that.
func fetchSecret(f func(*corev1.Secret) bool) matcher.GomegaMatcher {

	return WithTransform(func(secret *corev1.Secret) bool {

		k8sClient, _, err := utils.GetE2ETestKubeClientWithError()
		if err != nil {
			GinkgoWriter.Println(err)
			return false
		}

		err = k8sClient.Get(context.Background(), client.ObjectKeyFromObject(secret), secret)
		if err != nil {
			GinkgoWriter.Println(err)
			return false
		}

		return f(secret)

	}, BeTrue())

}
