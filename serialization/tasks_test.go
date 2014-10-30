package serialization_test

import (
	"net/url"

	"github.com/cloudfoundry-incubator/receptor"
	. "github.com/cloudfoundry-incubator/receptor/serialization"
	"github.com/cloudfoundry-incubator/runtime-schema/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task Serialization", func() {
	Describe("TaskToResponse", func() {
		var task models.Task

		BeforeEach(func() {
			task = models.Task{
				TaskGuid:   "the-task-guid",
				Domain:     "the-domain",
				RootFSPath: "the-rootfs-path",
				Actions: []models.ExecutorAction{
					{
						Action: models.UploadAction{
							From: "from",
							To:   "to",
						},
					},
				},
				Stack:     "the-stack",
				MemoryMB:  100,
				DiskMB:    100,
				CPUWeight: 50,
				Log: models.LogConfig{
					Guid:       "the-log-config-guid",
					SourceName: "the-source-name",
				},
				Annotation: "the-annotation",

				CreatedAt:     1234,
				FailureReason: "the-failure-reason",
				Failed:        true,
				Result:        "the-result",
				State:         models.TaskStateInvalid,
			}
		})

		It("serializes the state", func() {
			EXPECTED_STATE_MAP := map[models.TaskState]string{
				models.TaskStateInvalid:   "INVALID",
				models.TaskStatePending:   "PENDING",
				models.TaskStateClaimed:   "CLAIMED",
				models.TaskStateRunning:   "RUNNING",
				models.TaskStateCompleted: "COMPLETED",
				models.TaskStateResolving: "RESOLVING",
			}

			for modelState, jsonState := range EXPECTED_STATE_MAP {
				task.State = modelState
				Ω(TaskToResponse(task).State).Should(Equal(jsonState))
			}
		})

		It("serializes the task's fields", func() {
			actualResponse := TaskToResponse(task)

			expectedResponse := receptor.TaskResponse{
				TaskGuid:   "the-task-guid",
				Domain:     "the-domain",
				RootFSPath: "the-rootfs-path",
				Actions: []models.ExecutorAction{
					{
						Action: models.UploadAction{
							From: "from",
							To:   "to",
						},
					},
				},
				Stack:     "the-stack",
				MemoryMB:  100,
				DiskMB:    100,
				CPUWeight: 50,
				Log: models.LogConfig{
					Guid:       "the-log-config-guid",
					SourceName: "the-source-name",
				},
				Annotation: "the-annotation",

				CreatedAt:     1234,
				FailureReason: "the-failure-reason",
				Failed:        true,
				Result:        "the-result",
				State:         receptor.TaskStateInvalid,
			}

			Ω(actualResponse).Should(Equal(expectedResponse))
		})

		Context("when the task has a CompletionCallbackURL", func() {
			BeforeEach(func() {
				task.CompletionCallbackURL = &url.URL{
					Scheme: "http",
					Host:   "example.com",
					Path:   "/the-path",
				}
			})

			It("serializes the completion callback URL", func() {
				Ω(TaskToResponse(task).CompletionCallbackURL).Should(Equal("http://example.com/the-path"))
			})
		})

		Context("when the task doesn't have a CompletionCallbackURL", func() {
			It("leaves the completion callback URL blank", func() {
				Ω(TaskToResponse(task).CompletionCallbackURL).Should(Equal(""))
			})
		})
	})

	Describe("TaskFromRequest", func() {
		var request receptor.CreateTaskRequest
		var expectedTask models.Task

		BeforeEach(func() {
			request = receptor.CreateTaskRequest{
				TaskGuid:   "the-task-guid",
				Domain:     "the-domain",
				RootFSPath: "the-rootfs-path",
				Actions: []models.ExecutorAction{
					{
						Action: &models.RunAction{
							Path: "the-path",
						},
					},
				},
				Stack:     "the-stack",
				MemoryMB:  100,
				DiskMB:    100,
				CPUWeight: 50,
				Log: models.LogConfig{
					Guid:       "the-log-config-guid",
					SourceName: "the-source-name",
				},
				ResultFile: "the/result/file",
				Annotation: "the-annotation",
			}

			expectedTask = models.Task{
				TaskGuid:   "the-task-guid",
				Domain:     "the-domain",
				RootFSPath: "the-rootfs-path",
				Actions: []models.ExecutorAction{
					{
						Action: &models.RunAction{
							Path: "the-path",
						},
					},
				},
				Stack:     "the-stack",
				MemoryMB:  100,
				DiskMB:    100,
				CPUWeight: 50,
				Log: models.LogConfig{
					Guid:       "the-log-config-guid",
					SourceName: "the-source-name",
				},
				ResultFile: "the/result/file",
				Annotation: "the-annotation",
			}
		})

		It("translates the request into a task model, preserving attributes", func() {
			actualTask, err := TaskFromRequest(request)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(actualTask).Should(Equal(expectedTask))
		})

		Context("when the request contains a parseable completion_callback_url", func() {
			BeforeEach(func() {
				request.CompletionCallbackURL = "http://stager.service.discovery.thing/endpoint"
			})

			It("parses the URL", func() {
				actualTask, err := TaskFromRequest(request)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(actualTask.CompletionCallbackURL).Should(Equal(&url.URL{
					Scheme: "http",
					Host:   "stager.service.discovery.thing",
					Path:   "/endpoint",
				}))
			})
		})

		Context("when the request contains an unparseable completion_callback_url", func() {
			BeforeEach(func() {
				request.CompletionCallbackURL = "ಠ_ಠ"
			})

			It("errors", func() {
				_, err := TaskFromRequest(request)
				Ω(err).Should(HaveOccurred())
			})
		})
	})
})