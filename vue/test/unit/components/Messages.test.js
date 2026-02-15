import { mount } from '@vue/test-utils'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import Messages from '@/components/Messages.vue'
import { createPinia } from 'pinia'

// Mock the external dependencies
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

describe('Messages.vue', () => {
  let wrapper
  let pinia

  const mockMessages = [
    {
      messageID: 'msg-1',
      content: 'Hello, this is a test message',
      userID: 'user-1',
      userName: 'TestUser1',
      channelID: 'channel-1',
      parentID: null,
      createdAt: '2025-02-01T10:00:00Z',
      updatedAt: '2025-02-01T10:00:00Z',
      isEdited: false,
      isDeleted: false,
      reactions: [],
      mentions: [],
      attachments: []
    },
    {
      messageID: 'msg-2',
      content: 'This is a reply message',
      userID: 'user-2',
      userName: 'TestUser2',
      channelID: 'channel-1',
      parentID: 'msg-1',
      createdAt: '2025-02-01T10:05:00Z',
      updatedAt: '2025-02-01T10:05:00Z',
      isEdited: false,
      isDeleted: false,
      reactions: [{ emoji: '👍', userIDs: ['user-1'] }],
      mentions: [],
      attachments: []
    }
  ]

  const mockChannel = {
    channelID: 'channel-1',
    channelName: 'Test Channel'
  }

  beforeEach(() => {
    pinia = createPinia()
    
    wrapper = mount(Messages, {
      global: {
        plugins: [pinia],
        stubs: {
          'EditBox': true,
          'EmojiModal': true,
          'EmojiedModal': true
        }
      },
      props: {
        messages: mockMessages,
        channel: mockChannel,
        aliases: [],
        groups: [],
        copyable: true,
        messageID: null
      }
    })
  })

  it('renders messages correctly', () => {
    expect(wrapper.find('[data-testid="messages-container"]').exists()).toBe(true)
    const messageElements = wrapper.findAll('[data-testid="message-item"]')
    expect(messageElements).toHaveLength(2)
  })

  it('displays message content correctly', () => {
    const messageElements = wrapper.findAll('[data-testid="message-item"]')
    expect(messageElements[0].text()).toContain('Hello, this is a test message')
    expect(messageElements[1].text()).toContain('This is a reply message')
  })

  it('displays user names', () => {
    const messageElements = wrapper.findAll('[data-testid="message-item"]')
    expect(messageElements[0].text()).toContain('TestUser1')
    expect(messageElements[1].text()).toContain('TestUser2')
  })

  it('displays timestamps', () => {
    const timeElements = wrapper.findAll('[data-testid="message-time"]')
    expect(timeElements).toHaveLength(2)
    expect(timeElements[0].text()).toBe('12:00')
    expect(timeElements[1].text()).toBe('12:00')
  })

  it('shows reactions when present', () => {
    const reactionElements = wrapper.findAll('[data-testid="message-reactions"]')
    expect(reactionElements[0].exists()).toBe(false) // First message has no reactions
    expect(reactionElements[1].exists()).toBe(true) // Second message has reactions
  })

  it('shows edited indicator for edited messages', async () => {
    const editedMessages = [
      {
        ...mockMessages[0],
        isEdited: true
      }
    ]

    await wrapper.setProps({ messages: editedMessages })
    
    const editedIndicator = wrapper.find('[data-testid="edited-indicator"]')
    expect(editedIndicator.exists()).toBe(true)
  })

  it('opens emoji modal when emoji button clicked', async () => {
    const emojiButton = wrapper.find('[data-testid="emoji-button"]')
    await emojiButton.trigger('click')
    
    // Check if the emoji modal function was called
    expect(wrapper.vm.openEmoji).toHaveBeenCalled()
  })

  it('opens emojied modal when emojied button clicked', async () => {
    const emojiedButton = wrapper.find('[data-testid="emojied-button"]')
    await emojiedButton.trigger('click')
    
    // Check if the emojied modal function was called
    expect(wrapper.vm.openEmojied).toHaveBeenCalled()
  })

  it('toggles edit when edit button clicked', async () => {
    const editButton = wrapper.find('[data-testid="edit-button"]')
    await editButton.trigger('click')
    
    // Check if the toggle edit function was called
    expect(wrapper.vm.toggleEdit).toHaveBeenCalled()
  })

  it('toggles bookmark when bookmark button clicked', async () => {
    const bookmarkButton = wrapper.find('[data-testid="bookmark-button"]')
    await bookmarkButton.trigger('click')
    
    // Check if the toggle bookmark function was called
    expect(wrapper.vm.toggleBookmark).toHaveBeenCalled()
  })

  it('displays thread messages correctly', async () => {
    const threadHead = {
      parentID: 'msg-1',
      messageID: 'thread-1'
    }

    await wrapper.setProps({ 
      messageID: 'thread-1',
      threadHead: threadHead
    })
    
    // Should show thread-specific UI
    expect(wrapper.find('[data-testid="thread-container"]').exists()).toBe(true)
  })

  it('handles copy functionality when copyable is true', async () => {
    const copyButton = wrapper.find('[data-testid="copy-button"]')
    await copyButton.trigger('click')
    
    // Should trigger copy functionality
    expect(wrapper.vm.copyMessage).toHaveBeenCalled()
  })

  it('does not show copy button when copyable is false', async () => {
    await wrapper.setProps({ copyable: false })
    
    const copyButton = wrapper.find('[data-testid="copy-button"]')
    expect(copyButton.exists()).toBe(false)
  })

  it('displays mentions correctly', async () => {
    const messageWithMention = {
      ...mockMessages[0],
      mentions: ['user-2'],
      content: '@TestUser2 Hello!'
    }

    await wrapper.setProps({ messages: [messageWithMention] })
    
    const mentionElement = wrapper.find('[data-testid="message-mention"]')
    expect(mentionElement.exists()).toBe(true)
    expect(mentionElement.text()).toContain('@TestUser2')
  })

  it('displays attachments when present', async () => {
    const messageWithAttachment = {
      ...mockMessages[0],
      attachments: [{
        type: 'image',
        url: '/test-image.jpg',
        name: 'test.jpg'
      }]
    }

    await wrapper.setProps({ messages: [messageWithAttachment] })
    
    const attachmentElement = wrapper.find('[data-testid="message-attachment"]')
    expect(attachmentElement.exists()).toBe(true)
  })

  it('applies correct CSS classes for different message types', async () => {
    const ownMessage = {
      ...mockMessages[0],
      userID: 'current-user'
    }

    await wrapper.setProps({ messages: [ownMessage] })
    
    const messageElement = wrapper.find('[data-testid="message-item"]')
    expect(messageElement.classes()).toContain('own-message')
  })
})
