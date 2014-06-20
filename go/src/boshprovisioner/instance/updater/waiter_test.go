package updater_test

import (
	"errors"
	"time"

	boshaction "bosh/agent/action"
	boshlog "bosh/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakebpagclient "boshprovisioner/agent/client/fakes"
	. "boshprovisioner/instance/updater"
)

var _ = Describe("Waiter", func() {
	var (
		sleptTimes  []time.Duration
		sleepFunc   func(d time.Duration)
		agentClient *fakebpagclient.FakeClient
		logger      boshlog.Logger
		waiter      Waiter
	)

	const (
		firstTimeGap      = 5000 * time.Millisecond
		subsequentTimeGap = 1000 * time.Millisecond
	)

	BeforeEach(func() {
		sleptTimes = []time.Duration{}
		sleepFunc = func(d time.Duration) { sleptTimes = append(sleptTimes, d) }

		agentClient = &fakebpagclient.FakeClient{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		waiter = NewWaiter(5000, 14000, sleepFunc, agentClient, logger)
	})

	Describe("Wait", func() {
		Context("when agent reports its state as 'running' after some time", func() {
			BeforeEach(func() {
				agentClient.GetStateStates = []boshaction.GetStateV1ApplySpec{
					boshaction.GetStateV1ApplySpec{JobState: "not-running"},
					boshaction.GetStateV1ApplySpec{JobState: "not-running"},
					boshaction.GetStateV1ApplySpec{JobState: "not-running"},
					boshaction.GetStateV1ApplySpec{JobState: "not-running"},
					boshaction.GetStateV1ApplySpec{JobState: "not-running"}, // 4
					boshaction.GetStateV1ApplySpec{JobState: "running"},
				}
			})

			It("waits for instance to become running checking every interval", func() {
				err := waiter.Wait()
				Expect(err).ToNot(HaveOccurred())

				Expect(sleptTimes).To(Equal([]time.Duration{
					firstTimeGap,
					subsequentTimeGap,
					subsequentTimeGap,
					subsequentTimeGap,
					subsequentTimeGap, // 4
					subsequentTimeGap,
				}))
			})
		})

		Context("when agent does not report its state as 'running' after some time", func() {
			BeforeEach(func() {
				agentClient.GetStateState = boshaction.GetStateV1ApplySpec{JobState: "not-running"}
			})

			It("return error after trying as many times as possible", func() {
				err := waiter.Wait()
				Expect(err).To(Equal(ErrNotRunning))

				Expect(sleptTimes).To(Equal([]time.Duration{
					firstTimeGap,
					subsequentTimeGap,
					subsequentTimeGap,
					subsequentTimeGap,
					subsequentTimeGap, // 4
					subsequentTimeGap,
					subsequentTimeGap,
					subsequentTimeGap,
					subsequentTimeGap, // 8
					subsequentTimeGap,
				}))
			})
		})

		Context("when agent state cannot be retrieved", func() {
			BeforeEach(func() {
				agentClient.GetStateErr = errors.New("fake-get-state-err")
			})

			It("returns error indicated failure to retrieve state", func() {
				err := waiter.Wait()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-get-state-err"))

				Expect(sleptTimes).To(Equal([]time.Duration{firstTimeGap}))
			})
		})

		Context("when watch time is starts with a 0", func() {
			BeforeEach(func() {
				waiter = NewWaiter(0, 14000, sleepFunc, agentClient, logger)
			})

			BeforeEach(func() {
				agentClient.GetStateStates = []boshaction.GetStateV1ApplySpec{
					boshaction.GetStateV1ApplySpec{JobState: "not-running"},
					boshaction.GetStateV1ApplySpec{JobState: "running"},
				}
			})

			It("immediately checks if instance is running", func() {
				err := waiter.Wait()
				Expect(err).ToNot(HaveOccurred())

				Expect(sleptTimes).To(Equal([]time.Duration{
					0 * time.Millisecond,
					subsequentTimeGap,
				}))
			})
		})

		Context("when watch time ends with a 0", func() {
			BeforeEach(func() {
				waiter = NewWaiter(0, 0, sleepFunc, agentClient, logger)
			})

			BeforeEach(func() {
				agentClient.GetStateStates = []boshaction.GetStateV1ApplySpec{
					boshaction.GetStateV1ApplySpec{JobState: "running"},
				}
			})

			It("immediately checks if instance is running", func() {
				err := waiter.Wait()
				Expect(err).ToNot(HaveOccurred())

				Expect(sleptTimes).To(Equal([]time.Duration{0 * time.Millisecond}))
			})
		})
	})
})
