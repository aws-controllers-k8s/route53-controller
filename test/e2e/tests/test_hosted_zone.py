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

import boto3
import logging
import time
from typing import Dict

import pytest

from acktest.k8s import resource as k8s
from acktest.k8s import condition
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_route53_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.bootstrap_resources import get_bootstrap_resources

RESOURCE_PLURAL = "hostedzones"

# Time to wait after modifying the CR for the status to change
MODIFY_WAIT_AFTER_SECONDS = 10

# Time to wait after the zone has changed status, for the CR to update
CHECK_STATUS_WAIT_SECONDS = 10

@pytest.fixture
def public_hosted_zone():
    zone_name = random_suffix_name("public-hosted-zone", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["ZONE_NAME"] = zone_name
    replacements["ZONE_DOMAIN"] = f"{zone_name}.ack.example.com."

    resource_data = load_route53_resource(
        "hosted_zone_public",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    # Create the k8s resource
    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
        zone_name, namespace="default",
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

@pytest.fixture
def private_hosted_zone():
    zone_name = random_suffix_name("private-hosted-zone", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["ZONE_NAME"] = zone_name
    replacements["ZONE_DOMAIN"] = f"{zone_name}.ack.example.com."
    replacements["VPC_ID"] = get_bootstrap_resources().HostedZoneVPC.vpc_id

    resource_data = load_route53_resource(
        "hosted_zone_private",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    # Create the k8s resource
    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
        zone_name, namespace="default",
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
class TestHostedZone:
    def test_create_delete_public(self, route53_client, public_hosted_zone):
        (ref, cr) = public_hosted_zone

        zone_id = cr["status"]["id"]

        assert zone_id

        try:
            aws_res = route53_client.get_hosted_zone(Id=zone_id)
            assert aws_res is not None
        except route53_client.exceptions.NoSuchHostedZone:
            pytest.fail(f"Could not find hosted zone with ID '{zone_id}' in Route53")

    def test_create_delete_private(self, route53_client, private_hosted_zone):
        (ref, cr) = private_hosted_zone

        zone_id = cr["status"]["id"]

        assert zone_id

        try:
            aws_res = route53_client.get_hosted_zone(Id=zone_id)
            assert aws_res is not None
        except route53_client.exceptions.NoSuchHostedZone:
            pytest.fail(f"Could not find hosted zone with ID '{zone_id}' in Route53")