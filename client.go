package radikocast

import "context"

const (
	envRadikoMail     = "RADIKO_MAIL"
	envRadikoPassword = "RADIKO_PASSWORD"
)

func getClient(_ context.Context, _ string) (interface{}, error) {
	return nil, nil
}
