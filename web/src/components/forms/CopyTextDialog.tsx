import { useState, useCallback } from 'react';

interface CopyTextDialogProps {
  text: string;
  onClose: () => void;
}

export function CopyTextDialog({ text, onClose }: CopyTextDialogProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback for environments without clipboard API
      const textarea = document.createElement('textarea');
      textarea.value = text;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  }, [text]);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={onClose}>
      <div
        className="bg-gray-800 border border-gray-700 rounded-lg shadow-xl max-w-lg w-full mx-4 p-4"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-medium text-gray-200">Copy to Clipboard</h3>
          <button
            type="button"
            onClick={onClose}
            className="text-gray-400 hover:text-gray-200 text-lg leading-none"
            aria-label="Close"
          >
            &times;
          </button>
        </div>
        <pre className="text-xs text-gray-300 bg-gray-900 rounded p-3 mb-3 whitespace-pre-wrap max-h-60 overflow-y-auto">
          {text}
        </pre>
        <div className="flex justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 text-sm border border-gray-600 text-gray-300 rounded hover:bg-gray-700"
          >
            Close
          </button>
          <button
            type="button"
            onClick={handleCopy}
            className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            {copied ? 'Copied!' : 'Copy'}
          </button>
        </div>
      </div>
    </div>
  );
}
