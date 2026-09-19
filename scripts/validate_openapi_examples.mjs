#!/usr/bin/env node
import fs from 'node:fs';

const bundlePath = process.argv[2];
if (!bundlePath) throw new Error('Usage: validate_openapi_examples.mjs <dereferenced-bundle.json>');
const api = JSON.parse(fs.readFileSync(bundlePath, 'utf8'));

function fail(message) {
  throw new Error(`OpenAPI contract check failed: ${message}`);
}

function isObject(value) {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function validate(value, schema, location) {
  if (!schema) return;
  if (schema.allOf) for (const part of schema.allOf) validate(value, part, location);
  if (schema.enum && !schema.enum.some((item) => item === value)) {
    fail(`${location} is not one of [${schema.enum.join(', ')}]`);
  }

  switch (schema.type) {
    case 'object': {
      if (!isObject(value)) fail(`${location} must be an object`);
      for (const name of schema.required ?? []) {
        if (!Object.hasOwn(value, name)) fail(`${location} is missing required property '${name}'`);
      }
      if (schema.additionalProperties === false) {
        for (const name of Object.keys(value)) {
          if (!Object.hasOwn(schema.properties ?? {}, name)) {
            fail(`${location} contains undeclared property '${name}'`);
          }
        }
      }
      for (const [name, child] of Object.entries(schema.properties ?? {})) {
        if (Object.hasOwn(value, name)) validate(value[name], child, `${location}.${name}`);
      }
      break;
    }
    case 'array':
      if (!Array.isArray(value)) fail(`${location} must be an array`);
      value.forEach((item, index) => validate(item, schema.items, `${location}[${index}]`));
      break;
    case 'string':
      if (typeof value !== 'string') fail(`${location} must be a string`);
      if (schema.minLength !== undefined && value.length < schema.minLength) fail(`${location} is shorter than minLength`);
      if (schema.maxLength !== undefined && value.length > schema.maxLength) fail(`${location} is longer than maxLength`);
      if (schema.pattern && !(new RegExp(schema.pattern).test(value))) fail(`${location} does not match pattern ${schema.pattern}`);
      if (schema.format === 'email' && !/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(value)) fail(`${location} is not an email`);
      if (schema.format === 'date-time' && !/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$/.test(value)) fail(`${location} is not date-time`);
      if (schema.format === 'uri' && !/^[A-Za-z][A-Za-z0-9+.-]*:\/\//.test(value)) fail(`${location} is not a URI`);
      break;
    case 'integer':
      if (!Number.isInteger(value)) fail(`${location} must be an integer`);
      break;
    case 'number':
      if (typeof value !== 'number' || !Number.isFinite(value)) fail(`${location} must be a number`);
      break;
    case 'boolean':
      if (typeof value !== 'boolean') fail(`${location} must be boolean`);
      break;
  }
  if (schema.minProperties !== undefined && Object.keys(value).length < schema.minProperties) {
    fail(`${location} has fewer than minProperties`);
  }
  if (schema.minimum !== undefined && value < schema.minimum) fail(`${location} is below minimum`);
}

function validateContent(content, location) {
  let count = 0;
  for (const [mediaType, media] of Object.entries(content ?? {})) {
    if (Object.hasOwn(media, 'example')) {
      validate(media.example, media.schema, `${location} (${mediaType})`);
      count += 1;
    }
    for (const [name, example] of Object.entries(media.examples ?? {})) {
      validate(example.value, media.schema, `${location}/${name} (${mediaType})`);
      count += 1;
    }
  }
  return count;
}

if (!/^3\.1(?:\.|$)/.test(api.openapi)) fail(`expected OpenAPI 3.1+, got ${api.openapi}`);
for (const path of Object.keys(api.paths ?? {})) {
  if (!/^\/api\/v\d+\//.test(path)) fail(`path '${path}' is missing the /api/v<version>/ prefix`);
}

const channels = api.paths['/api/v1/channels'];
for (const method of ['get', 'post']) if (!channels?.[method]) fail(`/api/v1/channels is missing ${method}`);
const channel = api.paths['/api/v1/channels/{channelId}'];
for (const method of ['get', 'patch', 'delete']) if (!channel?.[method]) fail(`/api/v1/channels/{channelId} is missing ${method}`);

let partialPatchFound = false;
let exampleCount = 0;
for (const [path, pathItem] of Object.entries(api.paths ?? {})) {
  for (const [method, operation] of Object.entries(pathItem)) {
    if (!['get', 'post', 'put', 'patch', 'delete', 'head', 'options', 'trace'].includes(method)) continue;
    const responses = operation.responses ?? {};
    const responseStatuses = new Set(Object.keys(responses));
    if (!Object.keys(responses).some((status) => /^2\d\d$/.test(status))) {
      fail(`${method.toUpperCase()} ${path} has no successful 2xx response`);
    }
    if (!responseStatuses.has('500')) {
      fail(`${method.toUpperCase()} ${path} must document 500 Internal Server Error`);
    }
    if ((operation.requestBody || (operation.parameters ?? []).length > 0) && !responseStatuses.has('400')) {
      fail(`${method.toUpperCase()} ${path} must document 400 Bad Request for invalid input`);
    }
    const requestSchema = operation.requestBody?.content?.['application/json']?.schema;
    if (method === 'patch' && requestSchema?.minProperties !== undefined) {
      if (requestSchema.minProperties < 1) fail(`${method} ${path} has invalid minProperties`);
      partialPatchFound = true;
    }
    if (operation.requestBody) exampleCount += validateContent(operation.requestBody.content, `request ${method} ${path}`);
    for (const [status, response] of Object.entries(responses)) {
      if (status === '204' && response.content) {
        fail(`${method.toUpperCase()} ${path} has a response body for 204 No Content`);
      }
      if (/^[45]\d\d$/.test(status) && !response.content?.['application/json']?.schema) {
        fail(`${method.toUpperCase()} ${path} error response ${status} must have an application/json schema`);
      }
      for (const [mediaType, media] of Object.entries(response.content ?? {})) {
        if (!media.schema) fail(`${method.toUpperCase()} ${path} response ${status} (${mediaType}) has no schema`);
      }
      if (response.content) exampleCount += validateContent(response.content, `response ${status} ${method} ${path}`);
    }
  }
}
if (!partialPatchFound) fail('no PATCH operation with a partial-update schema was found');
if (exampleCount < 5) fail(`expected at least 5 schema-checked examples, found ${exampleCount}`);
console.log(`OpenAPI contract checks passed: ${Object.keys(api.paths).length} paths, ${exampleCount} schema-checked examples.`);
