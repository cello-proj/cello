#!/usr/bin/env node
import * as cdk from '@aws-cdk/core';
import { AppCdkTypescriptStack } from '../lib/app-cdk-typescript-stack';

const app = new cdk.App();
new AppCdkTypescriptStack(app, 'AppCdkTypescriptStack');
