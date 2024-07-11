import * as aws from '@pulumi/aws';

import { commonLabels, globalName } from '../../configuration';

/**
 * Creates an S3 bucket.
 *
 * @param {string} name the bucket name
 * @returns {aws.s3.Bucket} the bucket
 */
export const createS3Bucket = (name: string): aws.s3.Bucket =>
  new aws.s3.Bucket(`s3-bucket-${globalName}-${name}`, {
    serverSideEncryptionConfiguration: {
      rule: {
        applyServerSideEncryptionByDefault: {
          sseAlgorithm: 'AES256',
        },
      },
    },
    tags: commonLabels,
  });
