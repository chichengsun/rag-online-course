import { describe, it, expect } from 'vitest'

describe('示例测试', () => {
  it('应该通过基本断言', () => {
    expect(1 + 1).toBe(2)
  })

  it('应该正确比较字符串', () => {
    expect('hello' + ' ' + 'world').toBe('hello world')
  })
})
