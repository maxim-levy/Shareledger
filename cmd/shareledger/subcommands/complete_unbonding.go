package subcommands

import (
	"fmt"

	"github.com/sharering/shareledger/client"
	"github.com/spf13/cobra"
)

var CompleteUnbondingCmd = &cobra.Command{
	Use:   "complete_unbonding",
	Short: "complete unbonding",
	RunE:  completeUnbonding,
}


func init() {
	CompleteUnbondingCmd.Flags().StringVar(&nodeAddress, "client", "", "Node address to query info. Example: tcp://127.0.0.1:46657")
}

func completeUnbonding(cmd *cobra.Command, args []string) (err error) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			err = fmt.Errorf("Error in showing this masternode earning")
		}
	}()

	var context client.CoreContext
	if nodeAddress == "" {

		context = client.NewCoreContextFromConfig(config)

	} else {

		context = client.NewCoreContextFromConfigWithClient(config, nodeAddress)
	}

	err = context.CompleteUnbonding()
	if err != nil {
		return err
	}

	return nil

}
