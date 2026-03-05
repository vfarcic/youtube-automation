import { FieldLabel } from './FieldLabel';

interface DateInputProps {
  name: string;
  fieldName: string;
  value: string;
  onChange: (fieldName: string, value: string) => void;
  required?: boolean;
  helpText?: string;
  complete?: boolean;
}

export function DateInput({
  name,
  fieldName,
  value,
  onChange,
  required,
  helpText,
  complete,
}: DateInputProps) {
  return (
    <div>
      <FieldLabel name={name} required={required} helpText={helpText} htmlFor={fieldName} complete={complete} />
      <input
        id={fieldName}
        type="datetime-local"
        value={value}
        onChange={(e) => onChange(fieldName, e.target.value)}
        className="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
      />
    </div>
  );
}
