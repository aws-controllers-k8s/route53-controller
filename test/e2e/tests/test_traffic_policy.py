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

"""Integration tests for the Route53 TrafficPolicy resource
"""

import logging
import time

import pytest

from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_route53_resource
from e2e.replacement_values import REPLACEMENT_VALUES

RESOURCE_PLURAL = "trafficpolicies"

# Time to wait after modifying the CR for the status to change
MODIFY_WAIT_AFTER_SECONDS = 5

# Time to wait after the zone has changed status, for the CR to update
CHECK_STATUS_WAIT_SECONDS = 5

@pytest.fixture
def basic_traffic_policy():
    policy_name = random_suffix_name("traffic-policy", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["TRAFFIC_POLICY_NAME"] = policy_name

    resource_data = load_route53_resource(
        "traffic_policy",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    # Create the k8s resource
    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
        policy_name, namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert cr is not None
    assert k8s.get_resource_exists(ref)

    yield (ref, cr)

    # Try to delete, if doesn't already exist
    try:
        _, deleted = k8s.delete_custom_resource(ref, 3, 10)
        assert deleted
    except:
        pass

@service_marker
@pytest.mark.canary
class TestTrafficPolicy:
    def test_crud_basic(self, route53_client, basic_traffic_policy):
        (ref, cr) = basic_traffic_policy

        policy_name = cr["spec"]["name"]
        policy_id = cr["status"]["id"]
        policy_version = cr["status"]["version"]

        assert policy_name
        assert policy_id
        assert policy_version == 1

        try:
            aws_res = route53_client.get_traffic_policy(Id=policy_id, Version=policy_version)
            assert aws_res is not None
        except route53_client.exceptions.NoSuchTrafficPolicy:
            pytest.fail(f"Could not find traffic policy with ID '{policy_id}' in Route53")

        assert aws_res["TrafficPolicy"]["Type"] == "TXT"

        updated_document = """
            {
                "AWSPolicyFormatVersion": "2015-10-01",
                "RecordType": "TXT",
                "Endpoints": {
                    "endpoint-start": {
                        "Type": "value",
                        "Value": "\\"updatedtxtvalue\\""
                    }
                },
                "StartEndpoint": "endpoint-start"
            }
        """

        updates = {
            "spec": {
                "document": updated_document
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        cr = k8s.get_resource(ref)
        assert cr is not None

        policy_version = cr["status"]["version"]
        assert policy_version == 2

        try:
            aws_res = route53_client.get_traffic_policy(Id=policy_id, Version=policy_version)
            assert aws_res is not None
        except route53_client.exceptions.NoSuchTrafficPolicy:
            pytest.fail(f"Could not find traffic policy with ID '{policy_id}' in Route53")

        assert "updatedtxtvalue" in aws_res["TrafficPolicy"]["Document"]