import { FieldLabel } from './FieldLabel';

interface TextAreaProps {
  name: string;
  fieldName: string;
  value: string;
  onChange: (fieldName: string, value: string) => void;
  placeholder?: string;
  required?: boolean;
  helpText?: string;
  rows?: number;
  complete?: boolean;
}

export function TextArea({
  name,
  fieldName,
  value,
  onChange,
  placeholder,
  required,
  helpText,
  rows = 4,
  complete,
}: TextAreaProps) {
  return (
    <div>
      <FieldLabel name={name} required={required} helpText={helpText} htmlFor={fieldName} complete={complete} />
      <textarea
        id={fieldName}
        value={value}
        onChange={(e) => onChange(fieldName, e.target.value)}
        placeholder={placeholder}
        rows={rows}
        className="w-full border border-gray-600 bg-gray-800 text-gray-100 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500 resize-y placeholder-gray-500"
      />
    </div>
  );
}
