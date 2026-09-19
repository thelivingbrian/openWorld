const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');

test('world palettes drive static and animated canvas colors and reset between worlds', () => {
    const context = vm.createContext({
        window: { addEventListener() {} },
        document: { getElementById() { return null; } },
    });
    vm.runInContext(fs.readFileSync(`${__dirname}/assets/canvas.js`, 'utf8'), context);
    vm.runInContext(`applyCanvasPalette({ cssRules: [
        { selectorText: '.custom', style: { backgroundColor: 'rgba(80, 120, 160, 0.5)' } },
        { selectorText: '.green', style: { backgroundColor: 'rgb(12, 34, 56)' } },
        { selectorText: '.green-b', style: { borderColor: 'rgb(0, 0, 0)' } }
    ] });`, context);
    assert.equal(vm.runInContext('COLOR_MAP.custom', context), 'rgba(80, 120, 160, 0.5)');
    assert.equal(vm.runInContext('COLOR_MAP.green', context), 'rgb(12, 34, 56)');
    assert.equal(vm.runInContext('mixColor(COLOR_MAP.custom, COLOR_MAP.green, 0.5)', context), 'rgba(46, 77, 108, 0.75)');
    vm.runInContext('applyCanvasPalette({ cssRules: [] })', context);
    assert.equal(vm.runInContext('COLOR_MAP.custom', context), undefined);
    assert.equal(vm.runInContext('COLOR_MAP.green', context), 'rgb(32, 255, 60)');
});
