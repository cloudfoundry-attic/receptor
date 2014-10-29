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
				RootFSPath: "the-rootfs-path",
				CreatedAt:  1234,
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
			response := TaskToResponse(task)

			Ω(response.TaskGuid).Should(Equal("the-task-guid"))
			Ω(response.RootFSPath).Should(Equal("the-rootfs-path"))
			Ω(response.CreatedAt).Should(Equal(int64(1234)))
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

		BeforeEach(func() {
			request = receptor.CreateTaskRequest{
				TaskGuid:   "the-task-guid",
				Domain:     "the-domain",
				Stack:      "the-stack",
				RootFSPath: "the-rootfs-path",
				Actions: []models.ExecutorAction{
					{
						Action: &models.RunAction{
							Path: "the-path",
						},
					},
				},
			}
		})

		Context("when the request contains a completion_callback_url", func() {
			var task models.Task

			BeforeEach(func() {
				request.CompletionCallbackURL = "http://example.com/the-path"
			})

			JustBeforeEach(func() {
				var err error
				task, err = TaskFromRequest(request)
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("translates the request into a task model, preserving attributes", func() {
				Ω(task.TaskGuid).Should(Equal("the-task-guid"))
				Ω(task.Domain).Should(Equal("the-domain"))
				Ω(task.Stack).Should(Equal("the-stack"))
				Ω(task.RootFSPath).Should(Equal("the-rootfs-path"))
			})

			It("parses the URL", func() {
				Ω(task.CompletionCallbackURL).Should(Equal(&url.URL{
					Scheme: "http",
					Host:   "example.com",
					Path:   "/the-path",
				}))
			})
		})
	})
})
