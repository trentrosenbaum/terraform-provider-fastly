package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var vclMetadata = ServiceMetadata{
	ServiceTypeVCL,
}

// Ordering is important - stored is processing order
// Conditions need to be updated first, as they can be referenced by other
// configuration objects (Backends, Request Headers, etc)
var vclService = &BaseServiceDefinition{
	Metadata: vclMetadata,
	Attributes: []ServiceAttributeDefinition{
		NewServiceSettings(),
		NewServiceCondition(vclMetadata),
		NewServiceDomain(vclMetadata),
		NewServiceHealthCheck(vclMetadata),
		NewServiceBackend(vclMetadata),
		NewServiceDirector(vclMetadata),
		NewServiceHeader(vclMetadata),
		NewServiceGzip(vclMetadata),
		NewServiceS3Logging(vclMetadata),
		NewServicePaperTrail(vclMetadata),
		NewServiceSumologic(vclMetadata),
		NewServiceGCSLogging(vclMetadata),
		NewServiceBigQueryLogging(vclMetadata),
		NewServiceSyslog(vclMetadata),
		NewServiceLogentries(vclMetadata),
		NewServiceSplunk(vclMetadata),
		NewServiceBlobStorageLogging(vclMetadata),
		NewServiceHTTPSLogging(vclMetadata),
		NewServiceLoggingElasticSearch(vclMetadata),
		NewServiceLoggingFTP(vclMetadata),
		NewServiceLoggingSFTP(vclMetadata),
		NewServiceLoggingDatadog(vclMetadata),
		NewServiceLoggingLoggly(vclMetadata),
		NewServiceLoggingGooglePubSub(vclMetadata),
		NewServiceLoggingScalyr(vclMetadata),
		NewServiceLoggingNewRelic(vclMetadata),
		NewServiceLoggingKafka(vclMetadata),
		NewServiceLoggingHeroku(vclMetadata),
		NewServiceLoggingHoneycomb(vclMetadata),
		NewServiceLoggingLogshuttle(vclMetadata),
		NewServiceLoggingOpenstack(vclMetadata),
		NewServiceLoggingDigitalOcean(vclMetadata),
		NewServiceLoggingCloudfiles(vclMetadata),
		NewServiceResponseObject(vclMetadata),
		NewServiceRequestSetting(vclMetadata),
		NewServiceVCL(vclMetadata),
		NewServiceSnippet(vclMetadata),
		NewServiceDynamicSnippet(vclMetadata),
		NewServiceCacheSetting(vclMetadata),
		NewServiceACL(vclMetadata),
		NewServiceDictionary(vclMetadata),
	},
}

func resourceServiceV1() *schema.Resource {
	return resourceService(vclService)
}
