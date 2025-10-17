import { describe, it, expect } from 'vitest'
import {
  validateEmail,
  validateUrl,
  validateMemorySetting,
  validateJavaPath,
  validateDirectoryName,
  validateSettings,
  validateModpackId,
  validateInstanceId,
  sanitizeFileName,
  isValidMinecraftVersion,
} from './validation'
import { LauncherSettings } from '../types/launcher'

describe('validateEmail', () => {
  it('should validate correct email addresses', () => {
    const validEmails = [
      'user@example.com',
      'test.email+tag@domain.co.uk',
      'user123@test-domain.com',
      'a@b.co',
      'very.common@example.com',
    ]

    validEmails.forEach(email => {
      expect(validateEmail(email)).toBe(true)
    })
  })

  it('should reject invalid email addresses', () => {
    const invalidEmails = [
      '',
      'plainaddress',
      '@missing-local.com',
      'username@',
      'username@.com',
      'username@com',
      'username@domain..com',
      'username@domain.c',
      'username space@domain.com',
      'username@domain space.com',
      'username@domain,com',
    ]

    invalidEmails.forEach(email => {
      expect(validateEmail(email)).toBe(false)
    })
  })

  it('should handle edge cases', () => {
    expect(validateEmail('test@example.com')).toBe(true)
    expect(validateEmail('Test@EXAMPLE.COM')).toBe(true)
    expect(validateEmail('test@sub.example.com')).toBe(true)
  })
})

describe('validateUrl', () => {
  it('should validate correct URLs', () => {
    const validUrls = [
      'https://www.example.com',
      'http://localhost:8080',
      'https://api.github.com/repos/user/repo',
      'ftp://files.example.com/path',
      'https://example.com:443/path?query=value#fragment',
      'https://subdomain.example.co.uk/path',
    ]

    validUrls.forEach(url => {
      expect(validateUrl(url)).toBe(true)
    })
  })

  it('should reject invalid URLs', () => {
    const invalidUrls = [
      '',
      'not-a-url',
      '://missing-protocol',
      'http://',
      'https://',
      'javascript:alert("xss")',
      'data:text/plain,<script>alert("xss")</script>',
      'ht tp://invalid-spaces.com',
      'example.com', // Missing protocol
    ]

    invalidUrls.forEach(url => {
      expect(validateUrl(url)).toBe(false)
    })
  })

  it('should handle edge cases', () => {
    expect(validateUrl('https://example.com/path with spaces')).toBe(false)
    expect(validateUrl('http://192.168.1.1:3000')).toBe(true)
    expect(validateUrl('https://[2001:db8::1]')).toBe(true)
  })
})

describe('validateMemorySetting', () => {
  it('should validate memory within valid range', () => {
    const validMemory = [1024, 2048, 4096, 8192, 16384, 32768]

    validMemory.forEach(memory => {
      expect(validateMemorySetting(memory)).toBe(true)
    })
  })

  it('should reject memory outside valid range', () => {
    const invalidMemory = [0, 512, 1023, 32769, 65536, -1024]

    invalidMemory.forEach(memory => {
      expect(validateMemorySetting(memory)).toBe(false)
    })
  })

  it('should handle boundary values', () => {
    expect(validateMemorySetting(1024)).toBe(true) // Minimum
    expect(validateMemorySetting(32768)).toBe(true) // Maximum
    expect(validateMemorySetting(1023)).toBe(false) // Just below minimum
    expect(validateMemorySetting(32769)).toBe(false) // Just above maximum
  })
})

describe('validateJavaPath', () => {
  it('should validate correct Java paths', () => {
    const validPaths = [
      '/usr/bin/java',
      '/usr/lib/jvm/java-11-openjdk/bin/java',
      'C:\\Program Files\\Java\\jdk-17\\bin\\java.exe',
      'C:\\Program Files\\Java\\jdk-17\\bin\\java',
      '/opt/java/bin/java',
      '/home/user/.sdkman/candidates/java/current/bin/java',
    ]

    validPaths.forEach(path => {
      expect(validateJavaPath(path)).toBe(true)
    })
  })

  it('should reject incorrect Java paths', () => {
    const invalidPaths = [
      '',
      '/usr/bin/notjava',
      'C:\\Program Files\\Java\\jdk-17\\bin\\javac.exe',
      '/path/to/java.sh',
      'java',
      '/path/to/java/',
      'C:\\path\\to\\',
    ]

    invalidPaths.forEach(path => {
      expect(validateJavaPath(path)).toBe(false)
    })
  })
})

describe('validateDirectoryName', () => {
  it('should validate correct directory names', () => {
    const validNames = [
      'instances',
      'my-instances',
      'my_instances',
      'Instances123',
      'A',
      'Directory With Spaces',
      'unicode-ñame',
      'CamelCase',
    ]

    validNames.forEach(name => {
      expect(validateDirectoryName(name)).toBe(true)
    })
  })

  it('should reject invalid directory names', () => {
    const invalidNames = [
      '',
      'con', // Windows reserved
      'prn', // Windows reserved
      'aux', // Windows reserved
      'nul', // Windows reserved
      'dir<name',
      'dir>name',
      'dir:name',
      'dir"name',
      'dir|name',
      'dir?name',
      'dir*name',
      'dir/name',
      'dir\\name',
      'a'.repeat(256), // Too long
      'dir\tname',
      'dir\nname',
    ]

    invalidNames.forEach(name => {
      expect(validateDirectoryName(name)).toBe(false)
    })
  })

  it('should handle boundary cases', () => {
    expect(validateDirectoryName('a')).toBe(true) // Minimum length
    expect(validateDirectoryName('a'.repeat(255))).toBe(true) // Maximum length
    expect(validateDirectoryName('a'.repeat(256))).toBe(false) // Too long
  })
})

describe('validateSettings', () => {
  const createValidSettings = (): LauncherSettings => ({
    memoryMb: 4096,
    theme: 'dark',
    javaPath: '/usr/bin/java',
    prismPath: '/usr/bin/prism',
    instancesDir: '/home/user/instances',
    autoUpdate: true,
  })

  it('should validate correct settings', () => {
    const settings = createValidSettings()
    const errors = validateSettings(settings)
    expect(errors).toHaveLength(0)
  })

  it('should detect invalid memory settings', () => {
    const settings = createValidSettings()
    settings.memoryMb = 0

    const errors = validateSettings(settings)
    expect(errors).toContain('Memory allocation must be between 1GB and 32GB')
  })

  it('should detect invalid theme settings', () => {
    const settings = createValidSettings()
    settings.theme = 'invalid' as any

    const errors = validateSettings(settings)
    expect(errors).toContain('Invalid theme setting')
  })

  it('should detect invalid Java path', () => {
    const settings = createValidSettings()
    settings.javaPath = '/invalid/path'

    const errors = validateSettings(settings)
    expect(errors).toContain('Invalid Java path')
  })

  it('should detect invalid instances directory', () => {
    const settings = createValidSettings()
    settings.instancesDir = 'invalid<directory'

    const errors = validateSettings(settings)
    expect(errors).toContain('Invalid instances directory name')
  })

  it('should detect multiple validation errors', () => {
    const settings = createValidSettings()
    settings.memoryMb = 0
    settings.theme = 'invalid' as any
    settings.javaPath = '/invalid/path'

    const errors = validateSettings(settings)
    expect(errors.length).toBeGreaterThan(1)
    expect(errors).toContain('Memory allocation must be between 1GB and 32GB')
    expect(errors).toContain('Invalid theme setting')
    expect(errors).toContain('Invalid Java path')
  })

  it('should allow optional fields to be undefined', () => {
    const settings: Partial<LauncherSettings> = {
      memoryMb: 4096,
      theme: 'dark',
      autoUpdate: false,
    }

    const errors = validateSettings(settings as LauncherSettings)
    expect(errors).toHaveLength(0)
  })
})

describe('validateModpackId', () => {
  it('should validate correct modpack IDs', () => {
    const validIds = [
      'modpack',
      'modpack-123',
      'modpack_456',
      'Modpack-ABC_123',
      'a',
      'modpack-with-many-dashes_and_underscores-123',
    ]

    validIds.forEach(id => {
      expect(validateModpackId(id)).toBe(true)
    })
  })

  it('should reject invalid modpack IDs', () => {
    const invalidIds = [
      '',
      'modpack id', // Space
      'modpack.id', // Dot
      'modpack@id', // Special character
      'modpack#id',
      'modpack$id',
      'modpack%id',
      'modpack&id',
      'modpack*id',
      'modpack(id)',
      'modpack+id',
      'modpack=id',
      'modpack[id]',
      'modpack{id}',
      'modpack\\id',
      'modpack/id',
      'modpack:id',
      'modpack;id',
      'modpack"id',
      'modpack\'id',
      'modpack<id>',
      'modpack>id',
      'modpack,id',
      'a'.repeat(65), // Too long
    ]

    invalidIds.forEach(id => {
      expect(validateModpackId(id)).toBe(false)
    })
  })

  it('should handle boundary cases', () => {
    expect(validateModpackId('a')).toBe(true) // Minimum length
    expect(validateModpackId('a'.repeat(64))).toBe(true) // Maximum length
    expect(validateModpackId('a'.repeat(65))).toBe(false) // Too long
  })
})

describe('validateInstanceId', () => {
  it('should use the same validation as modpack ID', () => {
    const testId = 'test-instance-123'

    expect(validateInstanceId(testId)).toBe(validateModpackId(testId))
  })

  it('should validate correct instance IDs', () => {
    const validIds = [
      'instance',
      'instance-1',
      'instance_2',
      'Instance-A_1',
    ]

    validIds.forEach(id => {
      expect(validateInstanceId(id)).toBe(true)
    })
  })
})

describe('sanitizeFileName', () => {
  it('should replace invalid characters with underscores', () => {
    const testCases = [
      { input: 'file<name', expected: 'file_name' },
      { input: 'file>name', expected: 'file_name' },
      { input: 'file:name', expected: 'file_name' },
      { input: 'file"name', expected: 'file_name' },
      { input: 'file|name', expected: 'file_name' },
      { input: 'file?name', expected: 'file_name' },
      { input: 'file*name', expected: 'file_name' },
      { input: 'file/name', expected: 'file_name' },
      { input: 'file\\name', expected: 'file_name' },
    ]

    testCases.forEach(({ input, expected }) => {
      expect(sanitizeFileName(input)).toBe(expected)
    })
  })

  it('should replace multiple spaces with single underscores', () => {
    expect(sanitizeFileName('file   name')).toBe('file_name')
    expect(sanitizeFileName('file name  with   spaces')).toBe('file_name_with_spaces')
  })

  it('should truncate long names', () => {
    const longName = 'a'.repeat(300)
    const sanitized = sanitizeFileName(longName)
    expect(sanitized.length).toBeLessThanOrEqual(255)
    expect(sanitized.length).toBe(255)
  })

  it('should preserve valid characters', () => {
    const validNames = [
      'normal-file_name.txt',
      'File123.ext',
      'UPPERCASE.TXT',
      'mixed-Case_Name.ext',
    ]

    validNames.forEach(name => {
      expect(sanitizeFileName(name)).toBe(name)
    })
  })

  it('should handle empty input', () => {
    expect(sanitizeFileName('')).toBe('')
  })
})

describe('isValidMinecraftVersion', () => {
  it('should validate standard Minecraft versions', () => {
    const validVersions = [
      '1.20.1',
      '1.19.4',
      '1.18.2',
      '1.17.1',
      '1.16.5',
      '1.15.2',
      '1.14.4',
      '1.13.2',
      '1.12.2',
      '1.11.2',
      '1.10.2',
      '1.9.4',
      '1.8.9',
      '1.7.10',
      '1.6.4',
      '1.5.2',
      '1.4.7',
      '1.3.2',
      '1.2.5',
      '1.1',
      '1.0',
    ]

    validVersions.forEach(version => {
      expect(isValidMinecraftVersion(version)).toBe(true)
    })
  })

  it('should validate versions with release candidates', () => {
    const rcVersions = [
      '1.20.1-rc1',
      '1.20-pre1',
      '1.19.4-rc2',
      '1.19.3-pre3',
      '1.18.2-rc1',
    ]

    rcVersions.forEach(version => {
      expect(isValidMinecraftVersion(version)).toBe(true)
    })
  })

  it('should validate snapshot versions', () => {
    const snapshots = [
      '1.20.1-snapshot',
      '23w17a',
      '22w42a',
      '1.19-snapshot',
    ]

    snapshots.forEach(version => {
      expect(isValidMinecraftVersion(version)).toBe(true)
    })
  })

  it('should reject invalid versions', () => {
    const invalidVersions = [
      '',
      'version',
      '1.20.1.2', // Too many segments
      '.1.20.1', // Starts with dot
      '1.20.1.', // Ends with dot
      'v1.20.1', // Starts with v
      '1..20.1', // Double dot
      '1.20.', // Incomplete
      'a.b.c', // Non-numeric
      '1.20.1 beta', // Space in version
    ]

    invalidVersions.forEach(version => {
      expect(isValidMinecraftVersion(version)).toBe(false)
    })
  })

  it('should handle edge cases', () => {
    expect(isValidMinecraftVersion('1.0')).toBe(true)
    expect(isValidMinecraftVersion('2.0.1')).toBe(true)
    expect(isValidMinecraftVersion('10.15.20')).toBe(true)
  })
})