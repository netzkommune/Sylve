package cluster

import (
	"github.com/alchemillahq/sylve/internal/config"
	clusterServiceInterfaces "github.com/alchemillahq/sylve/internal/interfaces/services/cluster"
	"github.com/alchemillahq/sylve/pkg/utils"
)

func (s *Service) Detail() *clusterServiceInterfaces.Detail {
	nodeId, err := utils.GetSystemUUID()
	if err != nil {
		return nil
	}

	hostname, err := utils.GetSystemHostname()
	if err != nil {
		return nil
	}

	apiPort := config.ParsedConfig.Port

	return &clusterServiceInterfaces.Detail{
		NodeID:   nodeId,
		Hostname: hostname,
		APIPort:  apiPort,
	}
}
