# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

"""Helper functions for route53 tests
"""

class Route53Validator:
    def __init__(self, route53_client):
        self.route53_client = route53_client

    def list_tags_for_resources(self, resource_id: str, resource_type: str):
        resource_id = resource_id.replace('/' + resource_type + '/', '')
        try:
            res = self.route53_client.list_tags_for_resource(
                ResourceType=resource_type,
                ResourceId=resource_id)
            assert res is not None
            if len(res["ResourceTagSet"]) > 0:
                return res["ResourceTagSet"]
            assert False
        except self.route53_client.exceptions.ClientError as e:
            return None

    def assert_cidr_collection(self, cr, exists=True):
        res = None
        found = False
        cidr_collection_id = cr["status"].get("collection", None).get("id", None)
        cidr_collection_name = cr["spec"].get("name", None)
        cidr_collection_locations = cr["spec"].get("locations", None)

        if cidr_collection_id is None:
            return

        try:
            res = self.route53_client.list_cidr_collections()
            res_cidr_collections = [cidr_collection for cidr_collection in res["CidrCollections"] if cidr_collection["Id"] == cidr_collection_id]
            found = len(res_cidr_collections) == 1
        except self.route53_client.exceptions.ClientError:
            pass
        assert found is exists
        if exists and cidr_collection_id and cidr_collection_name:
            assert cidr_collection_name in str(res_cidr_collections)

        try:
            res = self.route53_client.list_cidr_blocks(CollectionId=cidr_collection_id)
        except self.route53_client.exceptions.ClientError:
            pass
        if exists and cidr_collection_locations is not None:
            for location in cidr_collection_locations:
                assert location["locationName"] in str(res)
                assert location["cidrList"][0] in str(res)

    def assert_hosted_zone(self, hosted_zone_id: str, exists=True):
        found = False
        try:
            res = self.route53_client.get_hosted_zone(Id=hosted_zone_id)
            found = len(res["HostedZone"]) > 0
        except self.route53_client.exceptions.ClientError:
            pass
        assert found is exists
    
    def assert_health_check(self, cr, exists=True):
        res = None
        found = False
        health_check_id = cr["status"].get("id", None)
        ip_address = cr["spec"].get("healthCheckConfig", {}).get("ipAddress", None)
        
        try:
            res = self.route53_client.get_health_check(HealthCheckId=health_check_id)
            found = len(res["HealthCheck"]) > 0
        except self.route53_client.exceptions.ClientError:
            pass
        assert found is exists
        if exists and ip_address:
            assert ip_address in str(res)

    def assert_record_set(self, cr, domain, exists=True):
        res = None
        found = False
        ip_address = cr["spec"]["resourceRecords"][0]["value"] if "resourceRecords" in cr["spec"].keys() else None

        dnsName = ""
        if "name" in cr["spec"].keys():
            dnsName += cr["spec"]["name"] + "."
        dnsName += domain

        try:
            res = self.route53_client.list_resource_record_sets(
                HostedZoneId=cr["spec"]["hostedZoneID"],
                StartRecordName=dnsName,
                StartRecordType=cr["spec"]["recordType"]
            )
            found = len(res) > 0
        except self.route53_client.exceptions.ClientError:
            pass

        assert found is exists
        if exists and ip_address:
            assert ip_address in str(res)
