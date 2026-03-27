import { computeDynamicStyle, hasDynamicToken, stripDynamicTokens, DYNAMIC_COLOR_MAP } from './dynamic-tile';

describe('dynamic-tile', () => {
  describe('hasDynamicToken', () => {
    it('returns false for undefined or empty input', () => {
      expect(hasDynamicToken(undefined)).toBe(false);
      expect(hasDynamicToken('')).toBe(false);
    });

    it('returns false for plain static class strings', () => {
      expect(hasDynamicToken('blue green-b thick')).toBe(false);
      expect(hasDynamicToken('wall')).toBe(false);
    });

    it('detects sparkle token', () => {
      expect(hasDynamicToken('sparkle')).toBe(true);
      expect(hasDynamicToken('blue sparkle green')).toBe(true);
    });

    it('detects rainbow token', () => {
      expect(hasDynamicToken('rainbow')).toBe(true);
      expect(hasDynamicToken('rainbow-b')).toBe(true);
      expect(hasDynamicToken('blue rainbow')).toBe(true);
    });

    it('detects water token', () => {
      expect(hasDynamicToken('water')).toBe(true);
      expect(hasDynamicToken('water blue')).toBe(true);
    });

    it('detects cycle(...) token', () => {
      expect(hasDynamicToken('cycle(blue,green)')).toBe(true);
      expect(hasDynamicToken('cycle-b(blue,green)')).toBe(true);
      expect(hasDynamicToken('thick cycle(red,blue)')).toBe(true);
    });
  });

  describe('stripDynamicTokens', () => {
    it('returns empty string for undefined or empty input', () => {
      expect(stripDynamicTokens(undefined)).toBe('');
      expect(stripDynamicTokens('')).toBe('');
    });

    it('passes through static tokens unchanged', () => {
      expect(stripDynamicTokens('blue green-b thick')).toBe('blue green-b thick');
    });

    it('strips sparkle and leaves other tokens', () => {
      expect(stripDynamicTokens('sparkle blue')).toBe('blue');
      expect(stripDynamicTokens('sparkle')).toBe('');
    });

    it('strips rainbow and rainbow-b', () => {
      expect(stripDynamicTokens('rainbow blue')).toBe('blue');
      expect(stripDynamicTokens('rainbow-b blue')).toBe('blue');
    });

    it('strips water token', () => {
      expect(stripDynamicTokens('water')).toBe('');
      expect(stripDynamicTokens('water thick blue')).toBe('thick blue');
    });

    it('strips cycle(...) and cycle-b(...) tokens', () => {
      expect(stripDynamicTokens('cycle(blue,green)')).toBe('');
      expect(stripDynamicTokens('cycle-b(blue,green)')).toBe('');
      expect(stripDynamicTokens('thick cycle(red,blue) green')).toBe('thick green');
    });

    it('strips multiple dynamic tokens from the same string', () => {
      expect(stripDynamicTokens('rainbow sparkle water blue')).toBe('blue');
    });
  });

  describe('computeDynamicStyle', () => {
    it('returns empty object for undefined or empty input', () => {
      expect(computeDynamicStyle(undefined, 1000, 0, 0)).toEqual({});
      expect(computeDynamicStyle('', 1000, 0, 0)).toEqual({});
    });

    it('returns empty object for purely static tokens (no dynamic tokens)', () => {
      expect(computeDynamicStyle('blue green-b', 1000, 0, 0)).toEqual({});
      expect(computeDynamicStyle('wall thick', 1000, 2, 3)).toEqual({});
    });

    it('returns backgroundColor for cycle(colorA,colorB) with known colors', () => {
      const style = computeDynamicStyle('cycle(blue,green)', 0, 0, 0);
      expect(style['backgroundColor']).toBeDefined();
      expect(style['backgroundColor']).toMatch(/^rgb\(/);
    });

    it('returns borderColor and borderStyle for cycle-b(colorA,colorB)', () => {
      const style = computeDynamicStyle('cycle-b(blue,green)', 0, 0, 0);
      expect(style['borderColor']).toBeDefined();
      expect(style['borderColor']).toMatch(/^rgb\(/);
      expect(style['borderStyle']).toBe('solid');
      expect(style['borderWidth']).toBe('1px');
    });

    it('ignores cycle token when a color name is not in the map', () => {
      const style = computeDynamicStyle('cycle(unknown-color,blue)', 0, 0, 0);
      expect(style['backgroundColor']).toBeUndefined();
    });

    it('returns backgroundColor for rainbow token', () => {
      const style = computeDynamicStyle('rainbow', 0, 0, 0);
      expect(style['backgroundColor']).toBeDefined();
      expect(style['backgroundColor']).toMatch(/^rgb\(/);
    });

    it('returns borderColor for rainbow-b token', () => {
      const style = computeDynamicStyle('rainbow-b', 0, 0, 0);
      expect(style['borderColor']).toBeDefined();
      expect(style['borderStyle']).toBe('solid');
    });

    it('returns backgroundColor for water token', () => {
      const style = computeDynamicStyle('water', 0, 2, 3);
      expect(style['backgroundColor']).toBeDefined();
      expect(style['backgroundColor']).toMatch(/^rgb\(/);
    });

    it('returns filter for sparkle token', () => {
      const style = computeDynamicStyle('sparkle', 0, 0, 0);
      expect(style['filter']).toBeDefined();
      expect(style['filter']).toMatch(/^brightness\(/);
    });

    it('produces deterministic per-cell phase offsets for rainbow', () => {
      const timeMs = 5000;
      const styleA = computeDynamicStyle('rainbow', timeMs, 0, 0);
      const styleB = computeDynamicStyle('rainbow', timeMs, 1, 0);
      const styleC = computeDynamicStyle('rainbow', timeMs, 0, 1);
      // Different coordinates produce different colors
      expect(styleA['backgroundColor']).not.toBe(styleB['backgroundColor']);
      expect(styleA['backgroundColor']).not.toBe(styleC['backgroundColor']);
    });

    it('produces deterministic per-cell phase offsets for water', () => {
      const timeMs = 5000;
      const styleA = computeDynamicStyle('water', timeMs, 0, 0);
      const styleB = computeDynamicStyle('water', timeMs, 0, 1);
      expect(styleA['backgroundColor']).not.toBe(styleB['backgroundColor']);
    });

    it('cycle produces same color at same time and coordinates (deterministic)', () => {
      const style1 = computeDynamicStyle('cycle(blue,green)', 1234, 3, 7);
      const style2 = computeDynamicStyle('cycle(blue,green)', 1234, 3, 7);
      expect(style1['backgroundColor']).toBe(style2['backgroundColor']);
    });

    it('produces different colors as time advances for cycle token', () => {
      const styleAtT0 = computeDynamicStyle('cycle(blue,green)', 0, 0, 0);
      const styleAtT1 = computeDynamicStyle('cycle(blue,green)', 10000, 0, 0);
      expect(styleAtT0['backgroundColor']).not.toBe(styleAtT1['backgroundColor']);
    });

    it('handles mixed static and dynamic tokens', () => {
      const style = computeDynamicStyle('thick blue rainbow', 0, 0, 0);
      expect(style['backgroundColor']).toBeDefined();
      expect(Object.keys(style).length).toBeGreaterThan(0);
    });

    it('cycle color interpolates between the two named endpoint colors', () => {
      // At timeMs=0 with y=x=0: mix = 0.5 + 0.5*sin(0) = 0.5
      // At mix=0.5 the result is midpoint of blue and green
      const blue = DYNAMIC_COLOR_MAP['blue'];
      const green = DYNAMIC_COLOR_MAP['green'];
      const style = computeDynamicStyle('cycle(blue,green)', 0, 0, 0);
      expect(style['backgroundColor']).toBeDefined();
      // Result must be an rgb() string different from both pure endpoints
      expect(style['backgroundColor']).not.toBe(blue);
      expect(style['backgroundColor']).not.toBe(green);
      expect(style['backgroundColor']).toMatch(/^rgb\(/);
    });
  });
});
