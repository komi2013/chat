import { describe, it, expect, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useMessagesStore } from '@/stores/messages'

describe('Messages Store', () => {
  let store

  beforeEach(() => {
    setActivePinia(createPinia())
    store = useMessagesStore()
  })

  it('initializes with empty messages array', () => {
    expect(store.messages).toEqual([])
  })

  it('inserts new message', () => {
    const message = {
      messageID: 'msg-1',
      content: 'Test message',
      userID: 'user-1'
    }

    store.insert(message)

    expect(store.messages).toHaveLength(1)
    expect(store.messages[0]).toEqual(message)
  })

  it('does not insert duplicate message', () => {
    const message = {
      messageID: 'msg-1',
      content: 'Test message',
      userID: 'user-1'
    }

    store.insert(message)
    store.insert(message) // Duplicate

    expect(store.messages).toHaveLength(1)
  })

  it('unshifts message at specific position', () => {
    const message1 = { messageID: 'msg-1', content: 'Message 1' }
    const message2 = { messageID: 'msg-2', content: 'Message 2' }
    const message3 = { messageID: 'msg-3', content: 'Message 3' }

    store.insert(message1)
    store.insert(message2)
    store.unshift(message3, 1) // Insert at position 1

    expect(store.messages).toHaveLength(3)
    expect(store.messages[0].messageID).toBe('msg-1')
    expect(store.messages[1].messageID).toBe('msg-3')
    expect(store.messages[2].messageID).toBe('msg-2')
  })

  it('updates existing message', () => {
    const message = {
      messageID: 'msg-1',
      content: 'Original message',
      userID: 'user-1'
    }

    store.insert(message)
    
    const updatedMessage = {
      ...message,
      content: 'Updated message'
    }
    
    store.update(updatedMessage, 'msg-1')

    expect(store.messages[0].content).toBe('Updated message')
  })

  it('does not update non-existent message', () => {
    const message = {
      messageID: 'msg-1',
      content: 'Test message'
    }

    store.insert(message)
    store.update({ messageID: 'msg-2', content: 'Different' }, 'msg-2')

    expect(store.messages[0].content).toBe('Test message')
    expect(store.messages).toHaveLength(1)
  })

  it('deletes message by ID', () => {
    const message1 = { messageID: 'msg-1', content: 'Message 1' }
    const message2 = { messageID: 'msg-2', content: 'Message 2' }

    store.insert(message1)
    store.insert(message2)
    store.delete('msg-1')

    expect(store.messages).toHaveLength(1)
    expect(store.messages[0].messageID).toBe('msg-2')
  })

  it('deletes all messages', () => {
    const message1 = { messageID: 'msg-1', content: 'Message 1' }
    const message2 = { messageID: 'msg-2', content: 'Message 2' }

    store.insert(message1)
    store.insert(message2)
    store.deleteAll()

    expect(store.messages).toHaveLength(0)
  })

  it('checks current display for channel messages', () => {
    const message = {
      messageID: 'msg-1',
      channelID: 'channel-1',
      content: 'Channel message'
    }

    store.insert(message)

    expect(store.currentDisplay('channel-1')).toBe(true)
    expect(store.currentDisplay('channel-2')).toBe(false)
  })

  it('checks current display for thread messages', () => {
    const message = {
      messageID: 'msg-1',
      parentID: 'parent-1',
      content: 'Thread message'
    }

    store.insert(message)

    expect(store.currentDisplay('parent-1')).toBe(true)
    expect(store.currentDisplay('different-parent')).toBe(false)
  })

  it('updates single field of message', () => {
    const message = {
      messageID: 'msg-1',
      content: 'Original content',
      isEdited: false
    }

    store.insert(message)
    store.upOne('msg-1', 'isEdited', true)

    expect(store.messages[0].isEdited).toBe(true)
    expect(store.messages[0].content).toBe('Original content')
  })

  it('does not update field for non-existent message', () => {
    const message = {
      messageID: 'msg-1',
      content: 'Test message',
      isEdited: false
    }

    store.insert(message)
    store.upOne('msg-2', 'isEdited', true)

    expect(store.messages[0].isEdited).toBe(false)
  })
})
