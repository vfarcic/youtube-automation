import { FieldLabel } from './FieldLabel';

interface ToggleProps {
  name: string;
  fieldName: string;
  value: boolean;
  onChange: (fieldName: string, value: boolean) => void;
  helpText?: string;
  complete?: boolean;
}

export function Toggle({ name, fieldName, value, onChange, helpText, complete }: ToggleProps) {
  return (
    <div className="flex items-center justify-between py-1">
      <FieldLabel name={name} helpText={helpText} htmlFor={fieldName} complete={complete} />
      <button
        id={fieldName}
        type="button"
        role="switch"
        aria-checked={value}
        onClick={() => onChange(fieldName, !value)}
        className={`relative inline-flex h-5 w-9 shrink-0 rounded-full border-2 border-transparent transition-colors ${value ? 'bg-blue-500' : 'bg-gray-300'}`}
      >
        <span
          className={`pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform transition-transform ${value ? 'translate-x-4' : 'translate-x-0'}`}
        />
      </button>
    </div>
  );
}
