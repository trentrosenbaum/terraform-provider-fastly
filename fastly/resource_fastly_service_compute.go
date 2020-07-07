package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var computeMetadata = ServiceMetadata{
	ServiceTypeCompute,
}

// Ordering is important - stored is processing order
// Some objects may need to be updated first, as they can be referenced by other
// configuration objects (Backends, Request Headers, etc).
var computeService = &BaseServiceDefinition{
	Metadata: computeMetadata,
	Attributes: []ServiceAttributeDefinition{
		NewServiceDomain(computeMetadata),
		NewServiceHealthCheck(computeMetadata),
		NewServiceBackend(computeMetadata),
		NewServiceS3Logging(computeMetadata),
		NewServicePaperTrail(computeMetadata),
		NewServiceSumologic(computeMetadata),
		NewServiceGCSLogging(computeMetadata),
		NewServiceBigQueryLogging(computeMetadata),
		NewServiceSyslog(computeMetadata),
		NewServiceLogentries(computeMetadata),
		NewServiceSplunk(computeMetadata),
		NewServiceBlobStorageLogging(computeMetadata),
		NewServiceHTTPSLogging(computeMetadata),
		NewServiceLoggingElasticSearch(computeMetadata),
		NewServiceLoggingFTP(computeMetadata),
		NewServiceLoggingSFTP(computeMetadata),
		NewServiceLoggingDatadog(computeMetadata),
		NewServiceLoggingLoggly(computeMetadata),
		NewServiceLoggingGooglePubSub(computeMetadata),
		NewServiceLoggingScalyr(computeMetadata),
		NewServiceLoggingNewRelic(computeMetadata),
		NewServiceLoggingKafka(computeMetadata),
		NewServiceLoggingHeroku(computeMetadata),
		NewServiceLoggingHoneycomb(computeMetadata),
		NewServiceLoggingLogshuttle(computeMetadata),
		NewServiceLoggingOpenstack(computeMetadata),
		NewServiceLoggingDigitalOcean(computeMetadata),
		NewServiceLoggingCloudfiles(computeMetadata),
		NewServicePackage(computeMetadata),
	},
}

func resourceServiceComputeV1() *schema.Resource {
	return resourceService(computeService)
}
