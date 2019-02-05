package openstack_test

import (
	"errors"
	"fmt"
	"sync"
	"time"

	. "github.com/cloudfoundry-incubator/bits-service/blobstores/openstack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteInParallel", func() {
	Context("names is empty", func() {
		It("doesn't call the deletionFunc", func() {
			errs := DeleteInParallel([]string{}, func(name string) error {
				defer GinkgoRecover()

				Fail("This function should not be called for an empty names slice")
				return nil
			})
			Expect(errs).To(BeEmpty())
		})
	})

	Context("names contains one element", func() {
		It("calls the deletionFunc once and returns no error", func() {
			errs := DeleteInParallel([]string{"foo"}, func(name string) error {
				println(name)
				return nil
			})
			Expect(errs).To(BeEmpty())
		})

		Context("deletionFunc returns an error", func() {
			It("returns the error as a result", func() {
				errs := DeleteInParallel([]string{"foo"}, func(name string) error {
					return errors.New("some error")
				})
				Expect(errs).To(ConsistOf(MatchError("some error")))
			})
		})
	})

	Context("names contains many elements where each item takes 10ms to delete ", func() {
		It("calls the deletionFunc for every item and finishes within 2s", func() {
			const numNames = 10000
			names := []string{}
			for i := 0; i < numNames; i++ {
				names = append(names, fmt.Sprintf("%v", i))
			}
			var errs []error
			var m sync.Mutex
			namesDeleted := make(map[string]bool)
			go func() {
				errs = DeleteInParallel(names, func(name string) error {
					time.Sleep(10 * time.Millisecond)
					m.Lock()
					defer m.Unlock()
					namesDeleted[name] = true
					return nil
				})
			}()
			Eventually(func() bool {
				for i := 0; i < numNames; i++ {
					m.Lock()
					if _, exists := namesDeleted[fmt.Sprintf("%v", i)]; !exists {
						m.Unlock()
						return false
					}
					m.Unlock()
				}
				return true
			}, "2s").Should(BeTrue())
			Expect(errs).To(BeEmpty())
		})

	})
})
