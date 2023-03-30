// Copyright (C) 2022 Specter Ops, Inc.
//
// This file is part of AzureHound.
//
// AzureHound is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// AzureHound is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/bloodhoundad/azurehound/client"
	"github.com/bloodhoundad/azurehound/enums"
	"github.com/bloodhoundad/azurehound/models"
	"github.com/spf13/cobra"
)

func init() {
	listRootCmd.AddCommand(listGroupsCmd)
}

var listGroupsCmd = &cobra.Command{
	Use:          "groups",
	Long:         "Lists Azure Active Directory Groups",
	Run:          listGroupsCmdImpl,
	SilenceUsage: true,
}

func listGroupsCmdImpl(cmd *cobra.Command, args []string) {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, os.Kill)
	defer gracefulShutdown(stop)

	log.V(1).Info("testing connections")
	azClient := connectAndCreateClient()
	log.Info("collecting azure active directory groups...")
	start := time.Now()
	stream := listGroups(ctx, azClient)
	outputStream(ctx, stream)
	duration := time.Since(start)
	log.Info("collection completed", "duration", duration.String())
}

func listGroups(ctx context.Context, client client.AzureClient) <-chan interface{} {
	out := make(chan interface{})

	go func() {
		defer close(out)
		count := 0
		for item := range client.ListAzureADGroups(ctx, "securityEnabled eq true", "", "", "", nil) {
			if item.Error != nil {
				log.Error(item.Error, "unable to continue processing groups")
				return
			} else {
				log.V(2).Info("found group", "group", item)
				count++
				group := models.Group{
					Group:      item.Ok,
					TenantId:   client.TenantInfo().TenantId,
					TenantName: client.TenantInfo().DisplayName,
				}
				out <- AzureWrapper{
					Kind: enums.KindAZGroup,
					Data: group,
				}
			}
		}
		log.Info("finished listing all groups", "count", count)
	}()

	return out
}
