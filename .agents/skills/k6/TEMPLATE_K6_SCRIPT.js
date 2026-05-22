// K6 Script Template for <MODULE> Workflow
import http from 'k6/http';
import { check, sleep } from 'k6';
import { authPatch, getCsrfToken } from '../auth-helper.js';
import config from '../config.js';

export const options = {
    vus: __VUS__,
    duration: '__DURATION__',
    thresholds: config.thresholds,
};

export default function () {
    // 1. Authenticate and get CSRF token
    // 2. Prepare data (dynamic if possible)
    // 3. Execute workflow steps (API calls)
    // 4. Check responses and handle errors
    // 5. Sleep or loop as needed
}

// Fill in workflow steps based on Gill Me session and module documentation.
