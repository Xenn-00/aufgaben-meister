package worker

import (
	"fmt"

	worker_handler "github.com/Xenn-00/aufgaben-meister/internal/worker/handlers"
	worker_task "github.com/Xenn-00/aufgaben-meister/internal/worker/tasks"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

func RegisterWorkerHandlers(mux *asynq.ServeMux, h *worker_handler.WorkerHander) {
	mux.HandleFunc(
		worker_task.TaskSendProjectInvitationEmail,
		h.WorkerInvitationEmailHandler(),
	)
	mux.HandleFunc(
		worker_task.TaskInvitationExpire,
		h.WorkerInvitationExpireHandler(),
	)
	mux.HandleFunc(
		worker_task.TaskOverdueAufgabenReminders,
		h.OverdueAufgaben(),
	)
	mux.HandleFunc(worker_task.TaskSendProjectProgressReminder, h.ReminderAufgaben())
	mux.HandleFunc(worker_task.TaskHandoverRequestNotifyMeister, h.HandoverRequestNotifyMeister())
}

func RegisterCronJobs(s *asynq.Scheduler) error {
	jobs := []struct {
		spec  string
		task  *asynq.Task
		queue string
		desc  string
	}{
		{
			spec:  "0 0 * * *",
			task:  asynq.NewTask(worker_task.TaskInvitationExpire, nil),
			queue: "low",
			desc:  "expire project invitations",
		},
		{
			spec:  "0 */6 * * *",
			task:  asynq.NewTask(worker_task.TaskOverdueAufgabenReminders, nil),
			queue: "low",
			desc:  "send task overdue reminder",
		},
		{
			spec:  "*/10 * * * *",
			task:  asynq.NewTask(worker_task.TaskSendProjectProgressReminder, nil),
			queue: "low",
			desc:  "send task progress reminder",
		},
	}

	for _, job := range jobs {
		if _, err := s.Register(job.spec, job.task, asynq.Queue(job.queue)); err != nil {
			return fmt.Errorf("register %s failed: %w", job.desc, err)
		}
		log.Info().Msgf("scheduled: %s", job.desc)
	}

	return nil
}
