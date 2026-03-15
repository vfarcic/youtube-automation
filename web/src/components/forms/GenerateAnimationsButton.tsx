import { useGenerateAnimations } from '../../api/hooks';

interface GenerateAnimationsButtonProps {
  category: string;
  videoName: string;
  onApply: (animations: string, timecodes: string) => void;
}

export function GenerateAnimationsButton({ category, videoName, onApply }: GenerateAnimationsButtonProps) {
  const mutation = useGenerateAnimations();

  const handleClick = () => {
    mutation.mutate(
      { name: videoName, category },
      {
        onSuccess: (data) => {
          // Format animations as bulleted list (matching CLI behavior)
          const animationsText = data.animations.map((a) => `- ${a}`).join('\n');

          // Build timecodes from sections (matching CLI behavior)
          let timecodesText = '';
          if (data.sections.length > 0) {
            const lines = ['00:00 FIXME:'];
            for (const section of data.sections) {
              const name = section.replace(/^Section: /, '');
              lines.push(`FIXME:FIXME ${name}`);
            }
            timecodesText = lines.join('\n');
          }

          onApply(animationsText, timecodesText);
        },
      },
    );
  };

  return (
    <div>
      <button
        type="button"
        onClick={handleClick}
        disabled={mutation.isPending}
        className="ml-2 px-2 py-0.5 text-xs bg-gray-700 text-gray-300 rounded hover:bg-gray-600 disabled:opacity-50"
      >
        {mutation.isPending ? 'Generating...' : 'Generate from Gist'}
      </button>
      {mutation.error && (
        <p className="text-xs text-red-400 mt-1">{mutation.error.message}</p>
      )}
    </div>
  );
}
