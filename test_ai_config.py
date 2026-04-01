#!/usr/bin/env python3
"""AI Configuration Persistence Smoke Test"""

import time
import json
from playwright.sync_api import sync_playwright, expect

BASE_URL = "http://localhost:5173"

def test_ai_config():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context()
        page = context.new_page()

        # Track console messages
        console_msgs = []
        page.on("console", lambda msg: console_msgs.append(f"[{msg.type}] {msg.text}"))

        print("=" * 60)
        print("1. SAVE AND RESTORE VERIFICATION")
        print("=" * 60)

        # Navigate to app
        page.goto(BASE_URL)
        page.wait_for_load_state('networkidle')
        time.sleep(2)

        # Take initial screenshot
        page.screenshot(path='/tmp/ai_test_01_initial.png')
        print("Screenshot: /tmp/ai_test_01_initial.png")

        # Click on Connections tab (sidebar)
        try:
            connections_btn = page.locator('text=连接').first
            if connections_btn.is_visible():
                connections_btn.click()
                time.sleep(1)
                print("Clicked 连接 tab")
        except Exception as e:
            print(f"Could not find connections tab: {e}")

        page.screenshot(path='/tmp/ai_test_02_connections.png')
        print("Screenshot: /tmp/ai_test_02_connections.png")

        # Click New Connection button
        try:
            new_btn = page.locator('button:has-text("新建"), button:has-text("New"), [class*="add"], [class*="create"]').first
            if new_btn.is_visible():
                new_btn.click()
                time.sleep(1)
                print("Clicked new connection button")
        except Exception as e:
            print(f"Could not find new button: {e}")

        page.screenshot(path='/tmp/ai_test_03_new_form.png')
        print("Screenshot: /tmp/ai_test_03_new_form.png")

        # Fill basic connection info
        try:
            # Connection name
            name_input = page.locator('input[placeholder*="名"], input[placeholder*="name"], #name').first
            if name_input.is_visible():
                name_input.fill("AI Test Connection")
                print("Filled connection name: AI Test Connection")
        except Exception as e:
            print(f"Name input error: {e}")

        # Select database type (MySQL)
        try:
            type_selector = page.locator('[data-testid*="type"], select, [class*="type"]').first
            if type_selector.is_visible():
                type_selector.click()
                time.sleep(0.5)
                mysql_option = page.locator('text=MySQL').first
                if mysql_option.is_visible():
                    mysql_option.click()
                    print("Selected MySQL")
        except Exception as e:
            print(f"Type selection error: {e}")

        # Fill host
        try:
            host_input = page.locator('input[placeholder*="host"], input[placeholder*="主机"], #host').first
            if host_input.is_visible():
                host_input.fill("localhost")
                print("Filled host: localhost")
        except Exception as e:
            print(f"Host input error: {e}")

        # Fill port
        try:
            port_input = page.locator('input[placeholder*="port"], input[placeholder*="端口"], #port, [type="number"]').first
            if port_input.is_visible():
                port_input.fill("3306")
                print("Filled port: 3306")
        except Exception as e:
            print(f"Port input error: {e}")

        # Fill username
        try:
            username_input = page.locator('input[placeholder*="user"], input[placeholder*="用户"], #username').first
            if username_input.is_visible():
                username_input.fill("root")
                print("Filled username: root")
        except Exception as e:
            print(f"Username input error: {e}")

        # Click AI Assistant tab
        try:
            ai_tab = page.locator('text=AI 助手, text=AI Assistant, [class*="ai"]').first
            if ai_tab.is_visible():
                ai_tab.click()
                time.sleep(1)
                print("Clicked AI Assistant tab")
        except Exception as e:
            print(f"AI tab click error: {e}")

        page.screenshot(path='/tmp/ai_test_04_ai_tab.png')
        print("Screenshot: /tmp/ai_test_04_ai_tab.png")

        # Fill AI configuration
        ai_config = {
            "name": "Test DeepSeek",
            "provider": "deepseek",
            "api_host": "https://api.deepseek.com",
            "api_endpoint": "/v1/chat/completions",
            "api_key": "sk-test-key-12345",
            "model": "deepseek-chat",
            "temperature": "0.7",
            "description": "Test AI Description",
            "language": "zh-CN"
        }

        # Try to fill AI form fields
        try:
            # AI Name
            ai_name = page.locator('input[placeholder*="AI名"], input[placeholder*="assistant name"]').first
            if ai_name.is_visible():
                ai_name.fill(ai_config["name"])
                print(f"Filled AI name: {ai_config['name']}")
        except Exception as e:
            print(f"AI name error: {e}")

        try:
            # Provider
            provider_select = page.locator('select, [class*="provider"]').first
            if provider_select.is_visible():
                provider_select.select_option(label="DeepSeek")
                print(f"Selected provider: DeepSeek")
        except Exception as e:
            print(f"Provider selection error: {e}")

        try:
            # API Host
            api_host = page.locator('input[placeholder*="api.host"], input[placeholder*="API地址"]').first
            if api_host.is_visible():
                api_host.fill(ai_config["api_host"])
                print(f"Filled API host: {ai_config['api_host']}")
        except Exception as e:
            print(f"API host error: {e}")

        try:
            # API Key
            api_key = page.locator('input[placeholder*="key"], input[type="password"]').first
            if api_key.is_visible():
                api_key.fill(ai_config["api_key"])
                print(f"Filled API key: {ai_config['api_key']}")
        except Exception as e:
            print(f"API key error: {e}")

        page.screenshot(path='/tmp/ai_test_05_ai_filled.png')
        print("Screenshot: /tmp/ai_test_05_ai_filled.png")

        # Click Save button
        try:
            save_btn = page.locator('button:has-text("保存"), button:has-text("Save"), button:has-text("确定")').first
            if save_btn.is_visible():
                save_btn.click()
                time.sleep(2)
                print("Clicked Save button")
        except Exception as e:
            print(f"Save button error: {e}")

        page.screenshot(path='/tmp/ai_test_06_after_save.png')
        print("Screenshot: /tmp/ai_test_06_after_save.png")

        # ========================================
        # Verify - Re-open the connection
        # ========================================
        print("\n" + "=" * 60)
        print("RE-OPENING CONNECTION TO VERIFY RESTORE")
        print("=" * 60)

        # Go back to connections list
        try:
            connections_btn = page.locator('text=连接').first
            if connections_btn.is_visible():
                connections_btn.click()
                time.sleep(1)
        except Exception as e:
            print(f"Navigation error: {e}")

        page.screenshot(path='/tmp/ai_test_07_back_list.png')
        print("Screenshot: /tmp/ai_test_07_back_list.png")

        # Click on the created connection to edit
        try:
            conn_item = page.locator('text=AI Test Connection').first
            if conn_item.is_visible():
                conn_item.click()
                time.sleep(1)
                print("Clicked on AI Test Connection")
        except Exception as e:
            print(f"Connection click error: {e}")

        # Click Edit button
        try:
            edit_btn = page.locator('button:has-text("编辑"), button:has-text("Edit"), [class*="edit"]').first
            if edit_btn.is_visible():
                edit_btn.click()
                time.sleep(1)
                print("Clicked Edit button")
        except Exception as e:
            print(f"Edit button error: {e}")

        page.screenshot(path='/tmp/ai_test_08_edit_form.png')
        print("Screenshot: /tmp/ai_test_08_edit_form.png")

        # Click AI tab again
        try:
            ai_tab = page.locator('text=AI 助手, text=AI Assistant').first
            if ai_tab.is_visible():
                ai_tab.click()
                time.sleep(1)
                print("Clicked AI tab in edit mode")
        except Exception as e:
            print(f"AI tab in edit error: {e}")

        page.screenshot(path='/tmp/ai_test_09_ai_restored.png')
        print("Screenshot: /tmp/ai_test_09_ai_restored.png")

        # Get the AI form values for verification
        print("\n--- AI Configuration Restore Check ---")
        restored_values = {}

        try:
            # Check various input values
            inputs = page.locator('input').all()
            print(f"Found {len(inputs)} input fields")

            # Try to get values
            for i, inp in enumerate(inputs):
                try:
                    value = inp.input_value()
                    placeholder = inp.get_attribute('placeholder') or ''
                    if value:
                        print(f"  Input {i} ({placeholder}): {value[:30]}...")
                        restored_values[placeholder] = value
                except:
                    pass
        except Exception as e:
            print(f"Could not read input values: {e}")

        # ========================================
        # AI Test Verification
        # ========================================
        print("\n" + "=" * 60)
        print("2. AI TEST VERIFICATION")
        print("=" * 60)

        # Look for Test AI button
        try:
            test_btn = page.locator('button:has-text("测试 AI"), button:has-text("Test AI"), button:has-text("测试")').first
            if test_btn.is_visible():
                print("Found Test AI button")

                # First test with correct config (will fail because key is fake)
                test_btn.click()
                time.sleep(5)  # Wait for API call

                page.screenshot(path='/tmp/ai_test_10_test_result.png')
                print("Screenshot: /tmp/ai_test_10_test_result.png")

                # Check for result message
                result_text = page.text_content('body')
                if '成功' in result_text or 'success' in result_text.lower():
                    print("Test result: SUCCESS")
                elif '失败' in result_text or 'error' in result_text.lower():
                    print("Test result: FAILED (expected with fake key)")
                else:
                    print("Test result: Check screenshot")
        except Exception as e:
            print(f"AI test button error: {e}")

        # ========================================
        # Summary
        # ========================================
        print("\n" + "=" * 60)
        print("SUMMARY")
        print("=" * 60)

        print(f"\nConsole messages captured: {len(console_msgs)}")
        for msg in console_msgs[-10:]:
            if 'error' in msg.lower() or 'ai' in msg.lower():
                print(f"  {msg}")

        print("\nScreenshots saved to /tmp/ai_test_*.png")

        browser.close()

if __name__ == "__main__":
    test_ai_config()
