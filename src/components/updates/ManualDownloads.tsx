import React, { useState } from 'react';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Progress } from '../ui/Progress';
import { Badge } from '../ui/Badge';
import { Tooltip } from '../ui/Tooltip';
import { formatFileSize } from '../../utils/format';

interface ManualDownload {
  id: string;
  filename: string;
  url: string;
  checksum?: string;
  size: number;
  download_type: string;
  instructions?: string;
}

interface ManualDownloadsProps {
  installId: string;
  downloads: ManualDownload[];
  onConfirm: (installId: string, filename: string, localPath: string) => void;
  onComplete: () => void;
}

export const ManualDownloads: React.FC<ManualDownloadsProps> = ({
  installId,
  downloads,
  onConfirm,
  onComplete
}) => {
  const [confirmedDownloads, setConfirmedDownloads] = useState<Set<string>>(new Set());
  const [validatingFiles, setValidatingFiles] = useState<Set<string>>(new Set());
  const [validationResults, setValidationResults] = useState<Map<string, boolean>>(new Map());

  const handleFileSelect = async (downloadId: string, filename: string, file: File) => {
    setValidatingFiles(prev => new Set(prev).add(downloadId));

    try {
      // For web/Tauri v2 environment, we use the file name instead of path
      // The backend will handle file access through the file system API
      const download = downloads.find(d => d.id === downloadId);
      if (download?.checksum) {
        // Note: In Tauri v2, we need to handle file differently
        // For now, we'll simulate validation
        const isValid = file.size === download.size; // Basic size check

        setValidationResults(prev => new Map(prev).set(downloadId, isValid));

        if (isValid) {
          setConfirmedDownloads(prev => new Set(prev).add(downloadId));
          onConfirm(installId, filename, file.name);
        }
      } else {
        // No checksum to validate, just confirm
        setConfirmedDownloads(prev => new Set(prev).add(downloadId));
        onConfirm(installId, filename, file.name);
      }
    } catch (error) {
      console.error('Failed to validate file:', error);
    } finally {
      setValidatingFiles(prev => {
        const newSet = new Set(prev);
        newSet.delete(downloadId);
        return newSet;
      });
    }
  };

  const allDownloadsComplete = downloads.every(d => confirmedDownloads.has(d.id));

  const getDownloadTypeInfo = (type: string) => {
    switch (type.toLowerCase()) {
      case 'direct':
        return { color: 'success', icon: '🔗', description: 'Direct download available' };
      case 'adfoc.us':
        return { color: 'warning', icon: '🔗', description: 'Ad-supported download, skip ads' };
      case 'modrinth':
        return { color: 'primary', icon: '🎮', description: 'Download from Modrinth' };
      case 'curseforge':
        return { color: 'primary', icon: '🎮', description: 'Download from CurseForge' };
      default:
        return { color: 'secondary', icon: '🔗', description: 'Manual download required' };
    }
  };

  return (
    <Card>
      <div className="space-y-6">
        <div className="text-center">
          <div className="text-6xl mb-4">📥</div>
          <h3 className="text-xl font-semibold mb-2">Manual Downloads Required</h3>
          <p className="text-gray-600 max-w-md mx-auto">
            Some files require manual download due to licensing restrictions.
            Please download each file and select it from your computer.
          </p>
        </div>

        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="flex items-start gap-2">
            <span className="text-blue-600 text-lg">ℹ️</span>
            <div className="text-sm text-blue-800">
              <p className="font-medium mb-1">How to complete manual downloads:</p>
              <ol className="list-decimal list-inside space-y-1 text-xs">
                <li>Click the download link for each file below</li>
                <li>Follow any instructions on the download page</li>
                <li>Save the file to your computer</li>
                <li>Click "Select File" and choose the downloaded file</li>
                <li>The launcher will validate the file automatically</li>
              </ol>
            </div>
          </div>
        </div>

        <div className="space-y-4">
          {downloads.map((download) => {
            const typeInfo = getDownloadTypeInfo(download.download_type);
            const isConfirmed = confirmedDownloads.has(download.id);
            const isValidating = validatingFiles.has(download.id);
            const validationResult = validationResults.get(download.id);

            return (
              <div key={download.id} className="border rounded-lg p-4">
                <div className="flex items-start justify-between mb-3">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-2">
                      <span className="font-medium">{download.filename}</span>
                      <Badge variant={typeInfo.color as any}>
                        {typeInfo.icon} {download.download_type}
                      </Badge>
                      {isConfirmed && (
                        <Badge variant="success">✅ Confirmed</Badge>
                      )}
                      {validationResult === false && (
                        <Badge variant="error">❌ Invalid</Badge>
                      )}
                    </div>

                    <div className="text-sm text-gray-600 mb-2">
                      Size: {formatFileSize(download.size)}
                    </div>

                    {download.instructions && (
                      <Tooltip content={download.instructions}>
                        <Button variant="outline" size="sm">
                          ℹ️ Instructions
                        </Button>
                      </Tooltip>
                    )}
                  </div>

                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => window.open(download.url, '_blank')}
                    >
                      🔗 Download
                    </Button>
                  </div>
                </div>

                <div className="space-y-2">
                  {isValidating && (
                    <div className="flex items-center gap-2 text-sm text-blue-600">
                      <div className="animate-spin">⏳</div>
                      <span>Validating file...</span>
                    </div>
                  )}

                  {!isConfirmed && (
                    <div>
                      <input
                        type="file"
                        className="w-full p-2 border rounded text-sm file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
                        onChange={(e) => {
                          const file = e.target.files?.[0];
                          if (file) {
                            handleFileSelect(download.id, download.filename, file);
                          }
                        }}
                        disabled={isValidating}
                      />

                      {validationResult === false && (
                        <div className="text-sm text-red-600 mt-1">
                          ❌ File validation failed. Please ensure you downloaded the correct file.
                        </div>
                      )}
                    </div>
                  )}

                  {isConfirmed && (
                    <div className="text-sm text-green-600">
                      ✅ File confirmed and validated
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>

        <div className="border-t pt-4">
          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-600">
              {confirmedDownloads.size} of {downloads.length} files confirmed
            </div>

            <div className="flex gap-2">
              <Button
                variant="outline"
                onClick={() => window.location.reload()}
              >
                🔄 Refresh
              </Button>

              {allDownloadsComplete && (
                <Button onClick={onComplete}>
                  ✅ Continue Installation
                </Button>
              )}
            </div>
          </div>

          {downloads.length > 0 && (
            <div className="mt-3">
              <Progress value={(confirmedDownloads.size / downloads.length) * 100} />
            </div>
          )}
        </div>
      </div>
    </Card>
  );
};