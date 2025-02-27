package loop

import (
	"context"

	"github.com/qdm12/deunhealth/internal/docker"
)

type Runner interface {
	Run(ctx context.Context) (err error)
}

func (l *Loop) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	existingUnhealthies, err := l.docker.GetUnhealthy(ctx)
	if err != nil {
		return err
	}
	for _, unhealthy := range existingUnhealthies {
		l.restartUnhealthy(ctx, unhealthy)
	}

	unhealthies := make(chan docker.UnhealthyContainer)
	unhealthyStreamCrashed := make(chan error)

	go l.docker.StreamUnhealthy(ctx, unhealthies, unhealthyStreamCrashed)

	for {
		select {
		case <-ctx.Done():
			<-unhealthyStreamCrashed
			close(unhealthyStreamCrashed)
			close(unhealthies)

			return ctx.Err()

		case err := <-unhealthyStreamCrashed:
			close(unhealthyStreamCrashed)
			close(unhealthies)

			return err

		case unhealthy := <-unhealthies:
			l.restartUnhealthy(ctx, unhealthy)
		}
	}
}

func (l *Loop) restartUnhealthy(ctx context.Context, unhealthy docker.UnhealthyContainer) {
	l.logger.Info("container " + unhealthy.Name +
		" (image " + unhealthy.Image + ") is unhealthy, restarting it...")
	err := l.docker.RestartContainer(ctx, unhealthy.Name)
	if err != nil {
		l.logger.Error("failed restarting container: " + err.Error())
	} else {
		l.logger.Info("container " + unhealthy.Name + " restarted successfully")
	}
}
