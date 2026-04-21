import { describe, expect, it } from 'vitest'
import { formatNumber } from './format-number'

describe('formatNumber', () => {
  it('formats integer with russian separators', () => {
    expect(formatNumber(12500)).toBe('12\u00A0500')
  })
})

