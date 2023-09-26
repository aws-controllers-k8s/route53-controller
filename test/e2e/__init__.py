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

import logging
import pytest
import time
from typing import Dict, Any
from pathlib import Path

from acktest.k8s import resource as k8s
from acktest.resources import load_resource_file

CRD_GROUP = "route53.services.k8s.aws"
CRD_VERSION = "v1alpha1"
SERVICE_NAME = "route53"

MODIFY_WAIT_AFTER_SECONDS = 10

# PyTest marker for the current service
service_marker = pytest.mark.service(arg=SERVICE_NAME)

bootstrap_directory = Path(__file__).parent
resource_directory = Path(__file__).parent / "resources"


def load_eks_resource(resource_name: str, additional_replacements: Dict[str, Any] = {}):
    """ Overrides the default `load_resource_file` to access the specific resources
    directory for the current service.
    """
    return load_resource_file(resource_directory, resource_name, additional_replacements=additional_replacements)


def get_route53_resource(ref):
    """
    Attempts to get a Route53 custom resource if it exists.
    """
    resource = None
    try:
        resource = k8s.get_resource(ref)
        assert resource
    except:
        pass
    return resource


def create_route53_resource(
    resource_plural,
    resource_name,
    spec_file,
    replacements,
    namespace="default",
):
    """
    Creates Route53 custom resources. The existence of the resource will be
    checked upon creation.
    """
    resource_data = load_eks_resource(
        spec_file,
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, resource_plural,
        resource_name, namespace,
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert cr is not None
    assert k8s.get_resource_exists(ref)

    return ref, cr


def delete_route53_resource(ref):
    """
    Attempts to delete a Route53 custom resource if it exists.
    """
    try:
        _, deleted = k8s.delete_custom_resource(ref, 3, 10)
        assert deleted
    except:
        pass


def patch_route53_resource(ref, updates):
    """
    Checks for the existence of a Route53 custom resource and subsequently
    patches the resource.
    """
    assert k8s.get_resource_exists(ref)

    res = k8s.patch_custom_resource(ref, updates)
    time.sleep(MODIFY_WAIT_AFTER_SECONDS)
    assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=5)

    return res
