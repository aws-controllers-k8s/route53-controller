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

"""Integration tests for the Route53 HostedZone resource
"""

import pytest
import time

from acktest import tags
from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import service_marker, create_route53_resource, delete_route53_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.bootstrap_resources import get_bootstrap_resources
from e2e.tests.helper import Route53Validator

RESOURCE_PLURAL = "hostedzones"

# Time to wait after modifying the CR for the status to change
MODIFY_WAIT_AFTER_SECONDS = 10

CREATE_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10

# Time to wait after the zone has changed status, for the CR to update
CHECK_STATUS_WAIT_SECONDS = 10

@pytest.fixture
def public_hosted_zone(request):
    zone_name = random_suffix_name("public-hosted-zone", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["ZONE_NAME"] = zone_name
    replacements["ZONE_DOMAIN"] = f"{zone_name}.ack.example.com."

    marker = request.node.get_closest_marker("resource_data")
    if marker is not None:
        data = marker.args[0]
        if 'tag_key' in data:
            replacements["TAG_KEY"] = data['tag_key']
        if 'tag_value' in data:
            replacements["TAG_VALUE"] = data['tag_value']

    ref, cr = create_route53_resource(
        "hostedzones",
        zone_name,
        "hosted_zone_public",
        replacements,
    )

    yield ref, cr

    delete_route53_resource(ref)

@pytest.fixture
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

    yield ref, cr

    delete_route53_resource(ref)

@service_marker
@pytest.mark.canary
class TestHostedZone:
    @pytest.mark.resource_data({'tag_key': 'key', 'tag_value': 'value'})
    def test_create_delete_public(self, route53_client, public_hosted_zone):
        ref, cr = public_hosted_zone

        zone_id = cr["status"]["id"]

        assert zone_id

        # Check hosted_zone exists in AWS
        route53_validator = Route53Validator(route53_client)
        route53_validator.assert_hosted_zone(zone_id)

    def test_create_delete_private(self, route53_client, private_hosted_zone):
        ref, cr = private_hosted_zone

        zone_id = cr["status"]["id"]

        assert zone_id

        # Check hosted_zone exists in AWS
        route53_validator = Route53Validator(route53_client)
        route53_validator.assert_hosted_zone(zone_id)
    @pytest.mark.resource_data({'tag_key': 'key', 'tag_value': 'value'})
    def test_delegation_set(self, route53_client, public_hosted_zone):
        ref, cr = public_hosted_zone

        time.sleep(CREATE_WAIT_AFTER_SECONDS)

        resource = k8s.get_resource(ref)
        resource_id = cr["status"]["id"]

        assert resource_id


        # Check hosted_zone exists in AWS
        route53_validator = Route53Validator(route53_client)
        route53_validator.assert_hosted_zone(resource_id)

        assert resource["status"]["delegationSet"] is not None
        assert resource["status"]["delegationSet"]["callerReference"]
        assert resource["status"]["delegationSet"]["id"]
        assert len(resource["status"]["delegationSet"]["nameServers"]) > 0

    @pytest.mark.resource_data({'tag_key': 'initialtagkey', 'tag_value': 'initialtagvalue'})
    def test_crud_tags(self, route53_client, public_hosted_zone):
        ref, cr = public_hosted_zone

        resource = k8s.get_resource(ref)
        resource_id = cr["status"]["id"]

        time.sleep(CREATE_WAIT_AFTER_SECONDS)

        # Check hosted_zone exists in AWS
        route53_validator = Route53Validator(route53_client)
        route53_validator.assert_hosted_zone(resource_id)

        # Check system and user tags exist for hosted_zone resource
        hosted_zone = route53_validator.list_tags_for_resources(resource_id, "hostedzone")
        user_tags = {
            "initialtagkey": "initialtagvalue"
        }
        tags.assert_ack_system_tags(
            tags=hosted_zone["Tags"],
        )
        tags.assert_equal_without_ack_tags(
            expected=user_tags,
            actual=hosted_zone["Tags"],
        )

        # Only user tags should be present in Spec
        assert len(resource["spec"]["tags"]) == 1
        assert resource["spec"]["tags"][0]["key"] == "initialtagkey"
        assert resource["spec"]["tags"][0]["value"] == "initialtagvalue"

        # Update tags
        update_tags = [
                {
                    "key": "updatedtagkey",
                    "value": "updatedtagvalue",
                }
            ]

        # Patch the dhcpOptions, updating the tags with new pair
        updates = {
            "spec": {"tags": update_tags},
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        # Check resource synced successfully
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=5)

        # Check for updated user tags; system tags should persist
        hosted_zone = route53_validator.list_tags_for_resources(resource_id, "hostedzone")
        updated_tags = {
            "updatedtagkey": "updatedtagvalue"
        }
        tags.assert_ack_system_tags(
            tags=hosted_zone["Tags"],
        )
        tags.assert_equal_without_ack_tags(
            expected=updated_tags,
            actual=hosted_zone["Tags"],
        )

        # Only user tags should be present in Spec
        resource = k8s.get_resource(ref)
        assert len(resource["spec"]["tags"]) == 1
        assert resource["spec"]["tags"][0]["key"] == "updatedtagkey"
        assert resource["spec"]["tags"][0]["value"] == "updatedtagvalue"

        # Patch the dhcpOptions resource, deleting the tags
        updates = {
            "spec": {"tags": []},
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        # Check resource synced successfully
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=5)

        # Check for removed user tags; system tags should persist
        hosted_zone = route53_validator.list_tags_for_resources(resource_id, "hostedzone")
        tags.assert_ack_system_tags(
            tags=hosted_zone["Tags"],
        )
        tags.assert_equal_without_ack_tags(
            expected=[],
            actual=hosted_zone["Tags"],
        )

        # Check user tags are removed from Spec
        resource = k8s.get_resource(ref)
        assert len(resource["spec"]["tags"]) == 0

        # Delete k8s resource
        _, deleted = k8s.delete_custom_resource(ref)
        assert deleted is True

        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        # Check hosted_zone no longer exists in AWS
        route53_validator.assert_hosted_zone(resource_id, exists=False)
