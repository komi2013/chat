package com.chat.android.core.util

import android.net.Uri
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

@RunWith(RobolectricTestRunner::class)
class InvitationMapperTest {

    @Test
    fun `parseUri returns ProfileRoute for valid invitation URL`() {
        val uri = Uri.parse("https://chat.quigen.info/invitation/shop001/?code=abc123")
        val result = InvitationMapper.parseUri(uri)
        assertEquals("shop001", result?.id)
        assertEquals("abc123", result?.code)
    }

    @Test
    fun `parseUri returns ProfileRoute for valid profile URL`() {
        val uri = Uri.parse("https://chat.quigen.info/profile/shop001/?code=abc123")
        val result = InvitationMapper.parseUri(uri)
        assertEquals("shop001", result?.id)
        assertEquals("abc123", result?.code)
    }

    @Test
    fun `parseUri returns ProfileRoute for custom scheme`() {
        val uri = Uri.parse("chat://profile/shop001?code=abc123")
        val result = InvitationMapper.parseUri(uri)
        assertEquals("shop001", result?.id)
        assertEquals("abc123", result?.code)
    }

    @Test
    fun `parseUri returns null for invalid host`() {
        val uri = Uri.parse("https://wrong.host/invitation/shop001/?code=abc123")
        val result = InvitationMapper.parseUri(uri)
        assertNull(result)
    }

    @Test
    fun `parseUri returns null for missing code in invitation`() {
        val uri = Uri.parse("https://chat.quigen.info/invitation/shop001/")
        val result = InvitationMapper.parseUri(uri)
        assertNull(result)
    }

    @Test
    fun `parseReferrer returns ProfileRoute for valid referrer`() {
        val referrer = "m=shop001&code=abc123"
        val result = InvitationMapper.parseReferrer(referrer)
        assertEquals("shop001", result?.id)
        assertEquals("abc123", result?.code)
    }

    @Test
    fun `parseReferrer returns null for missing m`() {
        val referrer = "code=abc123"
        val result = InvitationMapper.parseReferrer(referrer)
        assertNull(result)
    }

    @Test
    fun `parseReferrer returns null for missing code`() {
        val referrer = "m=shop001"
        val result = InvitationMapper.parseReferrer(referrer)
        assertNull(result)
    }

    @Test
    fun `parseReferrer handles extra parameters`() {
        val referrer = "m=shop001&code=abc123&utm_source=google"
        val result = InvitationMapper.parseReferrer(referrer)
        assertEquals("shop001", result?.id)
        assertEquals("abc123", result?.code)
    }
}
