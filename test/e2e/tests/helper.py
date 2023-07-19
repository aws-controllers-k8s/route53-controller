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

import pytest

from typing import Union, Dict


class Route53Validator:
    def __init__(self, route53_client):
        self.route53_client = route53_client

    def list_tags_for_resources(self, resource_id: str, resource_type: str):
        resource_id = resource_id.replace('/' + resource_type + '/', '')
        try:
            aws_res = self.route53_client.list_tags_for_resource(
                ResourceType=resource_type,
                ResourceId=resource_id)
            assert aws_res is not None
            if len(aws_res["ResourceTagSet"]) > 0:
                return aws_res["ResourceTagSet"]
            assert False
        except self.route53_client.exceptions.ClientError as e:
            return None

    def assert_hosted_zone(self, hosted_zone_id: str, exists=True):
        res_found = False
        try:
            aws_res = self.route53_client.get_hosted_zone(Id=hosted_zone_id)
            res_found = len(aws_res["HostedZone"]) > 0
        except self.route53_client.exceptions.ClientError:
            pass
        assert res_found is exists
