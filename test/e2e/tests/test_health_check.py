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

"""Integration tests for the Route53 HealthCheck resource
"""

import pytest
import random
import socket
import struct
import time

from acktest import tags
from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import service_marker, get_route53_resource, create_route53_resource, delete_route53_resource, patch_route53_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.tests.helper import Route53Validator

RESOURCE_PLURAL = "healthchecks"

# Time to wait after modifying the CR for the status to change
MODIFY_WAIT_AFTER_SECONDS = 10

CREATE_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10

@pytest.fixture
def health_check(request):
    health_check_name = random_suffix_name("health-check", 32)
    ip_address = socket.inet_ntoa(struct.pack('>I', random.randint(1, 0xffffffff)))

    replacements = REPLACEMENT_VALUES.copy()
    replacements["HEALTH_CHECK_NAME"] = health_check_name
    replacements["IP_ADDR"] = ip_address
    replacements["TAG_KEY"] = "initialtagkey"
    replacements["TAG_VALUE"] = "initialtagvalue"

    ref, cr = create_route53_resource(
        "healthchecks",
        health_check_name,
        "health_check",
        replacements,
    )

    yield ref, cr

    delete_route53_resource(ref)


def patch_health_check(ref):
    updates = {
        "spec": {
            "healthCheckConfig": {
                "failureThreshold": 5,
            }
        }
    }
    patch_route53_resource(ref, updates)
    return get_route53_resource(ref)


@service_marker
@pytest.mark.canary
class TestHealthCheck:
    @pytest.mark.resource_data({'tag_key': 'key', 'tag_value': 'value'})
    def test_crud(self, route53_client, health_check):
        ref, cr = health_check

        health_check_id = cr["status"]["id"]

        assert health_check_id

        # Check health check exists in AWS
        route53_validator = Route53Validator(route53_client)
        route53_validator.assert_health_check(cr)

        # Update health check resource and check that the value is propagated to AWS
        updated = patch_health_check(ref)
        assert updated["spec"]["healthCheckConfig"]["failureThreshold"] != cr["spec"]["healthCheckConfig"]["failureThreshold"]

        # Check health check has been updated in AWS
        route53_validator.assert_health_check(updated)

        # Delete k8s resource
        _, deleted = k8s.delete_custom_resource(ref)
        assert deleted is True

        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        # Check health check no longer exists in AWS
        route53_validator.assert_health_check(cr, exists=False)


    @pytest.mark.resource_data({'tag_key': 'initialtagkey', 'tag_value': 'initialtagvalue'})
    def test_crud_tags(self, route53_client, health_check):
        ref, cr = health_check

        resource = k8s.get_resource(ref)
        resource_id = cr["status"]["id"]

        time.sleep(CREATE_WAIT_AFTER_SECONDS)

        # Check health check exists in AWS
        route53_validator = Route53Validator(route53_client)
        route53_validator.assert_health_check(cr)

        # Check system and user tags exist for health_check resource
        health_check = route53_validator.list_tags_for_resources(resource_id, "healthcheck")
        initial_tags = {
            "initialtagkey": "initialtagvalue"
        }

        tags.assert_ack_system_tags(
            tags=health_check["Tags"],
        )
        tags.assert_equal_without_ack_tags(
            expected=initial_tags,
            actual=health_check["Tags"],
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
        assert k8s.wait_on_condition(ref, "Ready", "True", wait_periods=5)

        # Check for updated user tags; system tags should persist
        health_check = route53_validator.list_tags_for_resources(resource_id, "healthcheck")
        updated_tags = {
            "updatedtagkey": "updatedtagvalue"
        }

        tags.assert_ack_system_tags(
            tags=health_check["Tags"],
        )
        tags.assert_equal_without_ack_tags(
            expected=updated_tags,
            actual=health_check["Tags"],
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
        assert k8s.wait_on_condition(ref, "Ready", "True", wait_periods=5)

        # Check for removed user tags; system tags should persist
        health_check = route53_validator.list_tags_for_resources(resource_id, "healthcheck")

        tags.assert_ack_system_tags(
            tags=health_check["Tags"],
        )
        tags.assert_equal_without_ack_tags(
            expected=[],
            actual=health_check["Tags"],
        )

        # Check user tags are removed from Spec
        resource = k8s.get_resource(ref)
        assert len(resource["spec"]["tags"]) == 0

        # Delete k8s resource
        _, deleted = k8s.delete_custom_resource(ref)
        assert deleted is True

        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        # Check health check no longer exists in AWS
        route53_validator.assert_health_check(cr, exists=False)
