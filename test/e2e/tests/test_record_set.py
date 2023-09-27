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

from acktest.resources import random_suffix_name
from e2e import service_marker, get_route53_resource, create_route53_resource, delete_route53_resource, patch_route53_resource
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
    replacements["SIMPLE_RECORD_DNS_NAME"] = domain
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
        assert cr["status"]["id"]

        # Check record set exists in AWS
        route53_validator.assert_record_set(cr)

        # Ensure that the status eventually switches from PENDING to INSYNC
        assert verify_status_insync(ref) is True

        # Update record set's resource records and check that the value is propagated to AWS
        updated = patch_record_set(ref)
        assert updated["spec"]["resourceRecords"][0]["value"] != cr["spec"]["resourceRecords"][0]["value"]

        # ChangeBatch ID should have changed
        assert updated["status"]["id"] != cr["status"]["id"]

        # Check record set has been updated in AWS
        route53_validator.assert_record_set(updated)
