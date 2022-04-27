package engine

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"

	"tinkoff-invest-bot/investapi"

	robotProto "tinkoff-invest-bot/robot/proto"
)

type investRobot struct {
	config *robotProto.RobotConfig
}

func New(config *robotProto.RobotConfig) *investRobot {
	return &investRobot{
		config: config,
	}
}

func (r *investRobot) Run(ctx context.Context) error {
	creds := oauth.NewOauthAccess(&oauth2.Token{AccessToken: r.config.GetAccessToken()})

	connection, err := grpc.Dial(
		r.config.GetTinkoffApiEndpoint(),
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
		grpc.WithPerRPCCredentials(creds),
	)
	if err != nil {
		return xerrors.Errorf("can't connect to api: %w", err)
	}
	defer func() {
		if err := connection.Close(); err != nil {
			fmt.Println("Can't close the connection")
		}
	}()

	client := investapi.NewSandboxServiceClient(connection)

	var header metadata.MD
	result, err := client.GetSandboxAccounts(ctx, &investapi.GetAccountsRequest{}, grpc.Header(&header))
	if err != nil {
		return xerrors.Errorf("error while processing request: %w", err)
	}

	fmt.Printf("x-tracking-id: %v\n", header["x-tracking-id"])
	fmt.Printf("Server responce: %v\n", result)

	fmt.Println("investRobot successfully finished")
	return nil
}
