import React from 'react';
import { Layout } from '../components/layout/Layout';
import { JavaStatus } from '../components/JavaStatus';
import { PrismStatus } from '../components/PrismStatus';

export const ToolsPage: React.FC = () => {
  return (
    <Layout>
      <div className="container mx-auto px-4 py-6 max-w-7xl">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Tools Management</h1>
          <p className="text-gray-600">
            Manage Java installations and Prism Launcher for your Minecraft experience.
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Java Management */}
          <div className="space-y-4">
            <div className="flex items-center space-x-2">
              <div className="text-2xl">☕</div>
              <h2 className="text-xl font-semibold text-gray-900">Java Management</h2>
            </div>
            <JavaStatus minecraftVersion="1.20.1" />
          </div>

          {/* Prism Management */}
          <div className="space-y-4">
            <div className="flex items-center space-x-2">
              <div className="text-2xl">🔮</div>
              <h2 className="text-xl font-semibold text-gray-900">Prism Launcher</h2>
            </div>
            <PrismStatus />
          </div>
        </div>

        {/* System Information */}
        <div className="mt-8">
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
            <h3 className="text-lg font-semibold text-blue-900 mb-3">About These Tools</h3>
            <div className="space-y-3 text-sm text-blue-800">
              <div>
                <strong>Java:</strong> Minecraft requires specific Java versions depending on the game version.
                The launcher can automatically download and manage the required Java versions for you.
              </div>
              <div>
                <strong>Prism Launcher:</strong> A custom Minecraft launcher that helps manage instances,
                modpacks, and mods. It provides a better experience than the vanilla launcher with
                features like automatic mod installation and instance management.
              </div>
              <div className="mt-4 p-3 bg-blue-100 rounded">
                <strong>💡 Tip:</strong> These tools are automatically managed by the launcher when you
                install modpacks. You can also manually install or remove them using the controls above.
              </div>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
};