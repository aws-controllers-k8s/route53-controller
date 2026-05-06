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

"""Integration tests for the Route53 RecordSet resource
"""

import pytest
import random
import socket
import struct
import time

from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, get_route53_resource, create_route53_resource, delete_route53_resource, patch_route53_resource, load_eks_resource
from e2e.bootstrap_resources import get_bootstrap_resources
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.tests.helper import Route53Validator

STATUS_UPDATE_RETRY_COUNT = 4
STATUS_UPDATE_WAIT_TIME = 30

@pytest.fixture(scope="function")
def private_hosted_zone():
    zone_name = random_suffix_name("private-hosted-zone", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["ZONE_NAME"] = zone_name
    replacements["ZONE_DOMAIN"] = f"{zone_name}.ack.example.com."
    replacements["VPC_ID"] = get_bootstrap_resources().HostedZoneVPC.vpc_id

    ref, cr = create_route53_resource(
        "hostedzones",
        zone_name,
        "hosted_zone_private",
        replacements,
    )

    zone_id = cr["status"]["id"]
    domain = replacements["ZONE_DOMAIN"]
    yield zone_id, domain

    delete_route53_resource(ref)

@pytest.fixture(scope="function")
def simple_record_set(private_hosted_zone):
    zone_id, domain = private_hosted_zone
    parsed_zone_id = zone_id.split("/")[-1]
    ip_address = socket.inet_ntoa(struct.pack('>I', random.randint(1, 0xffffffff)))
    simple_record_name = random_suffix_name("simple-record-name", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["SIMPLE_RECORD_NAME"] = simple_record_name
    replacements["HOSTED_ZONE_ID"] = parsed_zone_id
    replacements["IP_ADDR"] = ip_address

    ref, cr = create_route53_resource(
        "recordsets",
        simple_record_name,
        "record_set_simple",
        replacements
    )

    yield ref, cr

    delete_route53_resource(ref)

@pytest.fixture(scope="function")
def fqdn_record_set(private_hosted_zone):
    zone_id, domain = private_hosted_zone
    parsed_zone_id = zone_id.split("/")[-1]
    ip_address = socket.inet_ntoa(struct.pack('>I', random.randint(1, 0xffffffff)))
    simple_record_name = random_suffix_name("fqdn-record-name", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["FQDN_K8S_NAME"] = simple_record_name
    replacements["FQDN_SPEC_NAME"] = simple_record_name + "." + domain
    replacements["HOSTED_ZONE_ID"] = parsed_zone_id
    replacements["IP_ADDR"] = ip_address

    ref, cr = create_route53_resource(
        "recordsets",
        simple_record_name,
        "record_set_fqdn",
        replacements
    )

    yield ref, cr

    delete_route53_resource(ref)

def status_id_exists(ref):
    for _ in range(STATUS_UPDATE_RETRY_COUNT):
        record = get_route53_resource(ref)
        if "id" in record["status"].keys() and record["status"]["id"]:
            return True
        time.sleep(STATUS_UPDATE_WAIT_TIME)
    return False

def verify_status_insync(ref):
    for _ in range(STATUS_UPDATE_RETRY_COUNT):
        record = get_route53_resource(ref)
        if record["status"]["status"] == "INSYNC":
            return True
        time.sleep(STATUS_UPDATE_WAIT_TIME)
    return False

def patch_record_set(ref):
    ip_address = socket.inet_ntoa(struct.pack('>I', random.randint(1, 0xffffffff)))
    updates = {
        "spec": {
            "resourceRecords": [
                {"value": ip_address}
            ]
        }
    }
    patch_route53_resource(ref, updates)
    return get_route53_resource(ref)

@service_marker
@pytest.mark.canary
class TestRecordSet:
    def test_crud_simple_record(self, route53_client, private_hosted_zone, simple_record_set):
        zone_id, domain = private_hosted_zone
        assert zone_id

        # Check hosted zone exists in AWS
        route53_validator = Route53Validator(route53_client)
        route53_validator.assert_hosted_zone(zone_id)

        ref, cr = simple_record_set
        assert status_id_exists(ref) is True

        # Check record set exists in AWS
        route53_validator.assert_record_set(cr, domain)

        # Ensure that the status eventually switches from PENDING to INSYNC
        assert verify_status_insync(ref) is True

        # Update record set's resource records and check that the value is propagated to AWS
        updated = patch_record_set(ref)
        assert updated["spec"]["resourceRecords"][0]["value"] != cr["spec"]["resourceRecords"][0]["value"]

        # ChangeBatch ID should have changed
        assert updated["status"]["id"] != cr["status"]["id"]

        # Check record set has been updated in AWS
        route53_validator.assert_record_set(updated, domain)


    def test_cd_fqdn_record(self, route53_client, private_hosted_zone, fqdn_record_set):
        zone_id, domain = private_hosted_zone
        assert zone_id

        # Check hosted zone exists in AWS
        route53_validator = Route53Validator(route53_client)
        route53_validator.assert_hosted_zone(zone_id)

        ref, cr = fqdn_record_set
        assert status_id_exists(ref) is True

        # Check record set exists in AWS
        route53_validator.assert_record_set(cr, domain)

        # Ensure that the status eventually switches from PENDING to INSYNC
        assert verify_status_insync(ref) is True


    def test_adopted_record_set_reaches_synced_state(self, route53_client, private_hosted_zone):
        """Test that adopting an existing RecordSet reaches ACK.ResourceSynced: True.

        This validates the fix for https://github.com/aws-controllers-k8s/community/issues/2861
        where adopted RecordSets would get stuck with ACK.ResourceSynced: False because
        no ChangeInfo.Id is set during adoption (no create/update is performed).
        """
        zone_id, domain = private_hosted_zone
        parsed_zone_id = zone_id.split("/")[-1]
        ip_address = socket.inet_ntoa(struct.pack('>I', random.randint(1, 0xffffffff)))
        record_dns_name = "adopt-test"
        full_dns_name = f"{record_dns_name}.{domain}"

        # Create the record set directly in AWS via boto3 (simulating a pre-existing resource)
        route53_client.change_resource_record_sets(
            HostedZoneId=parsed_zone_id,
            ChangeBatch={
                "Changes": [
                    {
                        "Action": "CREATE",
                        "ResourceRecordSet": {
                            "Name": full_dns_name,
                            "Type": "A",
                            "TTL": 300,
                            "ResourceRecords": [{"Value": ip_address}],
                        },
                    }
                ]
            },
        )

        ref = None
        try:
            # Now adopt the record set via ACK
            adopt_record_name = random_suffix_name("adopt-record", 32)
            replacements = REPLACEMENT_VALUES.copy()
            replacements["ADOPT_RECORD_NAME"] = adopt_record_name
            replacements["ADOPT_RECORD_DNS_NAME"] = record_dns_name
            replacements["HOSTED_ZONE_ID"] = parsed_zone_id
            replacements["IP_ADDR"] = ip_address

            resource_data = load_eks_resource(
                "record_set_adopt",
                additional_replacements=replacements,
            )

            ref = k8s.CustomResourceReference(
                CRD_GROUP, CRD_VERSION, "recordsets",
                adopt_record_name, namespace="default",
            )
            k8s.create_custom_resource(ref, resource_data)
            cr = k8s.wait_resource_consumed_by_controller(ref)
            assert cr is not None

            # The adopted resource should reach a synced state without a change ID
            assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=10)

            # Verify no change ID was set (adoption doesn't trigger create/update)
            record = get_route53_resource(ref)
            assert record["status"].get("id") is None
        finally:
            # Clean up the K8s resource if it was created
            if ref is not None:
                delete_route53_resource(ref)

            # Ensure the AWS record is removed even if the K8s delete fails
            try:
                route53_client.change_resource_record_sets(
                    HostedZoneId=parsed_zone_id,
                    ChangeBatch={
                        "Changes": [
                            {
                                "Action": "DELETE",
                                "ResourceRecordSet": {
                                    "Name": full_dns_name,
                                    "Type": "A",
                                    "TTL": 300,
                                    "ResourceRecords": [{"Value": ip_address}],
                                },
                            }
                        ]
                    },
                )
            except route53_client.exceptions.InvalidChangeBatch:
                # Record already deleted by the controller
                pass
