package clusterServiceInterfaces

import clusterModels "github.com/alchemillahq/sylve/internal/db/models/cluster"

type Storages struct {
	S3 []clusterModels.ClusterS3Config `json:"s3"`
}
