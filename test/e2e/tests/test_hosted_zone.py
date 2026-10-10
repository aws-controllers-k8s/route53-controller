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
from e2e import service_marker, CRD_GROUP, CRD_VERSION, create_route53_resource, delete_route53_resource, load_eks_resource
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

@pytest.fixture
def private_hosted_zone_multiple_vpcs():
    zone_name = random_suffix_name("private-hosted-zone", 32)

    bootstrap = get_bootstrap_resources()
    replacements = REPLACEMENT_VALUES.copy()
    replacements["ZONE_NAME"] = zone_name
    replacements["ZONE_DOMAIN"] = f"{zone_name}.ack.example.com."
    replacements["VPC_ID"] = bootstrap.HostedZoneVPC.vpc_id
    replacements["VPC2_ID"] = bootstrap.VPC2.vpc_id
    replacements["VPC3_ID"] = bootstrap.VPC3.vpc_id

    ref, cr = create_route53_resource(
        "hostedzones",
        zone_name,
        "hosted_zone_private_multiple_vpcs",
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

    def test_update_spec_vpc(self, route53_client, private_hosted_zone):
        """Updating spec.vpc should disassociate the old VPC and associate the new one."""
        ref, cr = private_hosted_zone

        zone_id = cr["status"]["id"]
        bootstrap = get_bootstrap_resources()

        assert zone_id

        # Wait for initial create to complete
        time.sleep(CREATE_WAIT_AFTER_SECONDS)
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=10)

        route53_validator = Route53Validator(route53_client)
        route53_validator.assert_hosted_zone(zone_id)

        # Confirm the initial VPC is associated
        route53_validator.assert_vpc_association(zone_id, bootstrap.HostedZoneVPC.vpc_id, exists=True)

        # Patch spec.vpc to VPC2
        updates = {
            "spec": {
                "vpc": {
                    "vpcID": bootstrap.VPC2.vpc_id,
                    "vpcRegion": bootstrap.VPC2.region,
                }
            }
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=10)

        # VPC2 should now be associated; HostedZoneVPC should be disassociated
        route53_validator.assert_vpc_association(zone_id, bootstrap.VPC2.vpc_id, exists=True)
        route53_validator.assert_vpc_association(zone_id, bootstrap.HostedZoneVPC.vpc_id, exists=False)

    def test_delete_with_multiple_vpcs(self, route53_client, private_hosted_zone_multiple_vpcs):
        """Deleting a hosted zone with multiple VPCs should succeed (pre-delete disassociation)."""
        ref, cr = private_hosted_zone_multiple_vpcs

        zone_id = cr["status"]["id"]
        bootstrap = get_bootstrap_resources()

        assert zone_id

        # VPC associations are synced during the create reconciliation cycle.
        # Wait for the resource to be fully synced before checking associations.
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=10)

        # Verify all VPCs are associated before delete
        route53_validator = Route53Validator(route53_client)
        route53_validator.assert_vpc_association(zone_id, bootstrap.VPC2.vpc_id, exists=True)
        route53_validator.assert_vpc_association(zone_id, bootstrap.VPC3.vpc_id, exists=True)

        # Delete while multiple VPCs are still associated — should not fail
        _, deleted = k8s.delete_custom_resource(ref)
        assert deleted is True

        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        route53_validator.assert_hosted_zone(zone_id, exists=False)

    def test_crud_vpcs(self, route53_client, private_hosted_zone_multiple_vpcs):
        """Test CRUD operations for a private hosted zone with multiple VPCs."""
        ref, cr = private_hosted_zone_multiple_vpcs

        zone_id = cr["status"]["id"]
        bootstrap = get_bootstrap_resources()

        assert zone_id

        # VPC associations are synced during the create reconciliation cycle.
        # Wait for the resource to be fully synced before checking associations.
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=10)

        # Check hosted zone exists and all VPCs are associated
        route53_validator = Route53Validator(route53_client)
        route53_validator.assert_hosted_zone(zone_id)
        route53_validator.assert_vpc_association(zone_id, bootstrap.VPC2.vpc_id, exists=True)
        route53_validator.assert_vpc_association(zone_id, bootstrap.VPC3.vpc_id, exists=True)

        # Remove VPC3, keep VPC2 (and primary HostedZoneVPC)
        updates = {
            "spec": {"vpcs": [
                {"vpcID": bootstrap.HostedZoneVPC.vpc_id, "vpcRegion": bootstrap.HostedZoneVPC.region},
                {"vpcID": bootstrap.VPC2.vpc_id, "vpcRegion": bootstrap.VPC2.region},
            ]},
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=5)

        # VPC2 should remain, VPC3 should be disassociated
        route53_validator.assert_vpc_association(zone_id, bootstrap.VPC2.vpc_id, exists=True)
        route53_validator.assert_vpc_association(zone_id, bootstrap.VPC3.vpc_id, exists=False)

        # Remove VPC2, keep only the primary
        updates = {
            "spec": {"vpcs": [
                {"vpcID": bootstrap.HostedZoneVPC.vpc_id, "vpcRegion": bootstrap.HostedZoneVPC.region},
            ]},
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=5)

        # Both VPC2 and VPC3 should now be disassociated
        route53_validator.assert_vpc_association(zone_id, bootstrap.VPC2.vpc_id, exists=False)
        route53_validator.assert_vpc_association(zone_id, bootstrap.VPC3.vpc_id, exists=False)

        # Delete the zone
        _, deleted = k8s.delete_custom_resource(ref)
        assert deleted is True

        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        route53_validator.assert_hosted_zone(zone_id, exists=False)

    def test_adoption_name_mismatch(self, route53_client):
        """Adopting a zone with a wrong spec.name should result in ACK.Terminal condition.

        Validates the fix for the bug where adopting a HostedZone with a wrong spec.name
        would silently succeed instead of returning a terminal error. After the fix,
        the controller returns a terminal error so the mismatch is visible in
        status.conditions rather than silently ignored.
        """
        suffix = random_suffix_name("", 8).lstrip("-")
        actual_zone_name = f"ack-adopt-test-{suffix}.io."
        wrong_zone_domain = f"wrong-domain-{suffix}.com."
        cr_name = f"adoption-mismatch-{suffix}"

        # Step 1: Create a real Route53 hosted zone directly via boto3 (not via ACK)
        create_resp = route53_client.create_hosted_zone(
            Name=actual_zone_name,
            CallerReference=f"ack-test-{suffix}",
        )
        zone_id = create_resp["HostedZone"]["Id"]  # e.g. /hostedzone/ABCDEF123456

        ref = None
        ref_correct = None
        try:
            # Step 2: Apply adoption CR with wrong spec.name pointing to the zone ID
            replacements = REPLACEMENT_VALUES.copy()
            replacements["ZONE_NAME"] = cr_name
            replacements["WRONG_ZONE_DOMAIN"] = wrong_zone_domain
            replacements["ZONE_ID"] = zone_id

            resource_data = load_eks_resource(
                "hosted_zone_adopt_name_mismatch",
                additional_replacements=replacements,
            )

            ref = k8s.CustomResourceReference(
                CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
                cr_name, namespace="default",
            )
            k8s.create_custom_resource(ref, resource_data)
            k8s.wait_resource_consumed_by_controller(ref)

            # Step 3: Wait for reconcile
            time.sleep(MODIFY_WAIT_AFTER_SECONDS)

            # Step 4: Assert ACK.Terminal condition is True
            assert k8s.wait_on_condition(
                ref, "ACK.Terminal", "True", wait_periods=5, period_length=5
            ), "Expected ACK.Terminal condition to be True due to spec.name mismatch"

            # Step 5: Assert the terminal message contains zone name mismatch info
            terminal_cond = k8s.get_resource_condition(ref, "ACK.Terminal")
            assert terminal_cond is not None
            msg = terminal_cond.get("message", "")
            assert wrong_zone_domain in msg or actual_zone_name in msg, (
                f"Terminal message should reference mismatched names, got: {msg}"
            )

            # Step 6: Delete the CR with wrong spec.name
            # spec.name is immutable (CRD CEL rule: self == oldSelf), so we must
            # delete the CR and re-create it with the correct spec.name.
            # We use deletion_policy=retain so the underlying AWS zone is NOT deleted.
            k8s.patch_custom_resource(ref, {
                "metadata": {
                    "annotations": {
                        "services.k8s.aws/deletion-policy": "retain"
                    }
                }
            })
            k8s.delete_custom_resource(ref)
            time.sleep(DELETE_WAIT_AFTER_SECONDS)

            # Step 7: Re-create the CR with the correct spec.name
            cr_name_correct = f"adoption-correct-{suffix}"
            replacements_correct = REPLACEMENT_VALUES.copy()
            replacements_correct["ZONE_NAME"] = cr_name_correct
            replacements_correct["WRONG_ZONE_DOMAIN"] = actual_zone_name  # correct name this time
            replacements_correct["ZONE_ID"] = zone_id

            resource_data_correct = load_eks_resource(
                "hosted_zone_adopt_name_mismatch",
                additional_replacements=replacements_correct,
            )

            ref_correct = k8s.CustomResourceReference(
                CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
                cr_name_correct, namespace="default",
            )
            k8s.create_custom_resource(ref_correct, resource_data_correct)
            k8s.wait_resource_consumed_by_controller(ref_correct)

            # Step 8: Wait for reconcile
            time.sleep(MODIFY_WAIT_AFTER_SECONDS)

            # Step 9: Assert ACK.Terminal is not set and ACK.ResourceSynced is True
            assert k8s.wait_on_condition(
                ref_correct, "ACK.ResourceSynced", "True", wait_periods=10, period_length=5
            ), "Expected ACK.ResourceSynced to be True after correct spec.name adoption"

            terminal_cond_after = k8s.get_resource_condition(ref_correct, "ACK.Terminal")
            assert terminal_cond_after is None or terminal_cond_after.get("status") != "True", (
                "ACK.Terminal should not be True after correct spec.name adoption"
            )

        finally:
            # Cleanup: delete both CRs with retain policy, then delete the AWS zone
            for r in [ref, ref_correct]:
                if r is None:
                    continue
                try:
                    k8s.patch_custom_resource(r, {
                        "metadata": {
                            "annotations": {
                                "services.k8s.aws/deletion-policy": "retain"
                            }
                        }
                    })
                except Exception:
                    pass
                try:
                    k8s.delete_custom_resource(r)
                    time.sleep(DELETE_WAIT_AFTER_SECONDS)
                except Exception:
                    pass

            # Delete the AWS zone
            try:
                route53_client.delete_hosted_zone(Id=zone_id)
            except Exception:
                pass
