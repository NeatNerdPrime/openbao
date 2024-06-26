/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

import { Factory } from 'ember-cli-mirage';

export default Factory.extend({
  algorithm: 'SHA1',
  digits: 6,
  issuer: 'OpenBao',
  key_size: 20,
  max_validation_attempts: 5,
  name: '', // returned but cannot be set at this time
  namespace_id: 'root',
  period: 30,
  qr_size: 200,
  skew: 1,
  type: 'totp',

  afterCreate(record) {
    if (record.name) {
      console.warn('Endpoint ignored these unrecognized parameters: [name]'); // eslint-disable-line
      record.name = '';
    }
  },
});
