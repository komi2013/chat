import { vi } from 'vitest'
import { config } from '@vue/test-utils'

// Global mocks
global.localStorage = {
  getItem: vi.fn((key) => {
    const mockData = {
      'myname': 'TestUser',
      'csrf': 'test-csrf-token'
    }
    return mockData[key] || null
  }),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn()
}

// Mock fetch API
global.fetch = vi.fn(() =>
  Promise.resolve({
    ok: true,
    status: 200,
    json: () => Promise.resolve({}),
    text: () => Promise.resolve('')
  })
)

// Mock IndexedDB
const mockIDB = {
  open: vi.fn(() => ({
    onsuccess: vi.fn(),
    onerror: vi.fn(),
    onupgradeneeded: vi.fn()
  }))
}
global.indexedDB = mockIDB

// Mock WebSocket
global.WebSocket = vi.fn(() => ({
  close: vi.fn(),
  send: vi.fn(),
  addEventListener: vi.fn(),
  removeEventListener: vi.fn()
}))

// Mock Service Worker
global.navigator = {
  serviceWorker: {
    addEventListener: vi.fn(),
    register: vi.fn(() => Promise.resolve())
  }
}

// Mock window functions
global.alert = vi.fn()
global.confirm = vi.fn(() => true)
global.prompt = vi.fn(() => 'test-input')

// Vue Test Utils global configuration
config.global.stubs = {
  'router-link': true,
  'router-view': true,
  'transition': true,
  'transition-group': true
}

// Mock external modules
vi.mock('@/my/emoji', () => ({
  isEmojiOpen: vi.fn(() => false),
  selectedMessageId: vi.fn(() => 'test-message-id'),
  openEmoji: vi.fn(),
  closeEmoji: vi.fn(),
  selectEmoji: vi.fn(),
  calcEmoji: vi.fn(),
  emojiPath: vi.fn(() => '/emoji/test.png'),
  isEmojiedOpen: vi.fn(() => false),
  openEmojied: vi.fn(),
  closeEmojied: vi.fn(),
  rotateEmoji: vi.fn()
}))

vi.mock('@/my/toggle.js', () => ({
  toggleEdit: vi.fn(),
  toggleBookmark: vi.fn()
}))

vi.mock('@/my/markdown', () => ({
  markdownToHtml: vi.fn((text) => `<p>${text}</p>`)
}))

vi.mock('@/pushReceive/pushReceive.js', () => ({
  pushReceive: vi.fn()
}))

vi.mock('@/my/advertisement', () => ({
  loadAdvertisements: vi.fn(() => Promise.resolve([]))
}))

// Mock time functions
global.timeFormat = vi.fn((date) => '12:00')
global.checkFaviconBadge = vi.fn()

// Mock error logging
global.sendErrorLog = vi.fn()
