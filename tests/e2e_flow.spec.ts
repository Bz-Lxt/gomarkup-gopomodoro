import { test, expect } from '@playwright/test'

const BASE = process.env.E2E_BASE || 'http://frontend'

test('critical path: login → board → timer resume → abort → burndown', async ({ page }) => {
  await page.goto(BASE + '/login')
  await expect(page.getByRole('heading', { name: '量化时间座舱' })).toBeVisible()
  await page.getByRole('button', { name: '进入看板' }).click()
  await expect(page.getByRole('heading', { name: '敏捷看板' })).toBeVisible({ timeout: 15000 })

  const stamp = Date.now().toString()
  await page.getByRole('button', { name: '+ 任务' }).click()
  await page.getByPlaceholder('标题 *').fill('E2E 网关压测 ' + stamp)
  await page.getByRole('button', { name: '保存' }).click()
  await expect(page.getByText('E2E 网关压测 ' + stamp)).toBeVisible()

  await page.getByText('E2E 网关压测 ' + stamp).locator('..').getByRole('button', { name: '开始专注' }).click()
  await expect(page.getByText(/FOCUS LOCKED|STANDBY|HOLDING/)).toBeVisible({ timeout: 10000 })
  const before = await page.locator('.font-mono.text-7xl').innerText()
  await page.reload()
  await expect(page.locator('.font-mono.text-7xl')).toBeVisible({ timeout: 10000 })
  const after = await page.locator('.font-mono.text-7xl').innerText()
  expect(after).toMatch(/\d{2}:\d{2}/)
  expect(before).toMatch(/\d{2}:\d{2}/)

  if (await page.getByRole('button', { name: '暂停' }).isVisible()) {
    await page.getByRole('button', { name: '暂停' }).click()
  }
  if (await page.getByRole('button', { name: '放弃' }).isVisible()) {
    await page.getByRole('button', { name: '放弃' }).click()
    await page.getByRole('button', { name: '确认放弃' }).click()
  }

  await page.goto(BASE + '/burndown')
  await expect(page.getByText('今日番茄')).toBeVisible({ timeout: 15000 })
  await expect(page.getByText('专注废弃率')).toBeVisible()
})
