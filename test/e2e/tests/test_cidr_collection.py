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
import time

from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import service_marker, get_route53_resource, create_route53_resource, delete_route53_resource, patch_route53_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.tests.helper import Route53Validator

CREATE_WAIT_AFTER_SECONDS = 5
UPDATE_WAIT_AFTER_SECONDS = 5
DELETE_WAIT_AFTER_SECONDS = 10

@pytest.fixture
def cidr_collection(request):
    cidr_collection_name = random_suffix_name("cidr-collection", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CIDR_COLLECTION_NAME"] = cidr_collection_name

    ref, cr = create_route53_resource(
        "cidrcollections",
        cidr_collection_name,
        "cidr_collection",
        replacements,
    )

    yield ref, cr

    delete_route53_resource(ref)

def patch_cidr_collection(ref):
    updates = {
        "spec": {
            "locations": [{
                "locationName": "location-new",
                "cidrList": ["192.168.100.0/24"],
            }]
        }
    }
    patch_route53_resource(ref, updates)
    return get_route53_resource(ref)


@service_marker
@pytest.mark.canary
class TestCidrCollection:
    def test_crud(self, route53_client, cidr_collection):
        ref, cr = cidr_collection

        cidr_collection_name = cr["spec"]["name"]

        assert cidr_collection_name

        time.sleep(CREATE_WAIT_AFTER_SECONDS)

        # Check cidr collection exists in AWS
        route53_validator = Route53Validator(route53_client)
        route53_validator.assert_cidr_collection(cr)

        # Update cidr collection resource
        updated = patch_cidr_collection(ref)
        assert updated["spec"]["locations"] != cr["spec"]["locations"]

        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        # Check cidr collection has been updated in AWS
        route53_validator.assert_cidr_collection(updated)

        # Delete k8s resource
        _, deleted = k8s.delete_custom_resource(ref)
        assert deleted is True

        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        # Check cidr collection no longer exists in AWS
        route53_validator.assert_cidr_collection(cr, exists=False)
