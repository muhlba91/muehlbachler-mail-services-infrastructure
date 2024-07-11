import * as aws from '@pulumi/aws';
import { Output } from '@pulumi/pulumi';

import { createAccessKey } from '../aws/iam/key';
import { createAWSUser } from '../aws/iam/user';
import { commonLabels } from '../configuration';

/**
 * Creates the SimpleLogin AWS user.
 *
 * @param {string} bucketArn the ARN of the bucket to allow access to
 * @returns {Output<aws.iam.AccessKey>} the access key
 */
export const createUser = (bucketArn: string): Output<aws.iam.AccessKey> => {
  const user = createAWSUser('simplelogin', {
    policies: [
      new aws.iam.Policy(
        'aws-policy-simplelogin',
        {
          policy: aws.iam
            .getPolicyDocument({
              statements: [
                {
                  effect: 'Allow',
                  actions: ['s3:*'],
                  resources: [bucketArn + '/*'],
                },
              ],
            })
            .then((doc) => doc.json),
          tags: commonLabels,
        },
        {},
      ),
    ],
  });

  return user.name.apply((name) => createAccessKey(name, user));
};
