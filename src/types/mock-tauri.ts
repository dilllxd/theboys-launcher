// Mock Tauri API for development
export const invoke = async (command: string, args?: any): Promise<any> => {
  // Mock responses for development
  switch (command) {
    case 'get_available_modpacks':
      return [
        {
          id: 'theboys-vanilla',
          displayName: 'TheBoys Vanilla',
          packUrl: 'https://example.com/theboys-vanilla.zip',
          instanceName: 'TheBoys-Vanilla',
          description: 'TheBoys modpack with vanilla Minecraft experience',
          default: true,
          version: '1.0.0',
          minecraftVersion: '1.20.1',
          modloader: 'fabric' as const,
          loaderVersion: '0.14.24'
        },
        {
          id: 'theboys-enhanced',
          displayName: 'TheBoys Enhanced',
          packUrl: 'https://example.com/theboys-enhanced.zip',
          instanceName: 'TheBoys-Enhanced',
          description: 'TheBoys modpack with enhanced gameplay and visual improvements',
          default: false,
          version: '2.1.0',
          minecraftVersion: '1.20.1',
          modloader: 'forge' as const,
          loaderVersion: '47.3.0'
        }
      ];

    case 'get_installed_modpacks':
      return [];

    case 'get_default_modpack':
      return {
        id: 'theboys-vanilla',
        displayName: 'TheBoys Vanilla',
        packUrl: 'https://example.com/theboys-vanilla.zip',
        instanceName: 'TheBoys-Vanilla',
        description: 'TheBoys modpack with vanilla Minecraft experience',
        default: true,
        version: '1.0.0',
        minecraftVersion: '1.20.1',
        modloader: 'fabric' as const,
        loaderVersion: '0.14.24'
      };

    case 'check_all_modpack_updates':
      return [];

    case 'download_modpack':
      return `download-${Date.now()}`;

    default:
      console.log(`Mock invoke: ${command}`, args);
      return null;
  }
};