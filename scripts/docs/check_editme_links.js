const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');

const source = fs.readFileSync('docs/_layouts/page.html', 'utf8');
const script = source.match(/<script>([\s\S]*?)<\/script>/)[1];
for (const [version, branch] of Object.entries({latest: 'main', v3: 'main', v2: '2', 'v1.2': '1.2', 'pr-123': 'pr-123'})) {
    let href;
    const link = {prop: (_, value) => { href = value; return link; }, removeClass: () => link};
    vm.runInNewContext(script, {
        document: {},
        getDocVersionFromPage: () => version,
        $: selector => selector === '#editme_link' ? link : {ready: callback => callback()},
    });
    assert.ok(href.includes(`/blob/${branch}/docs/`), `${version}: ${href}`);
}
console.log('Documentation Source links select the matching branch');
