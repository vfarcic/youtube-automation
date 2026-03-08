import { FieldLabel } from './FieldLabel';

interface NumberInputProps {
  name: string;
  fieldName: string;
  value: number;
  onChange: (fieldName: string, value: number) => void;
  required?: boolean;
  helpText?: string;
  min?: number;
  max?: number;
  complete?: boolean;
}

export function NumberInput({
  name,
  fieldName,
  value,
  onChange,
  required,
  helpText,
  min,
  max,
  complete,
}: NumberInputProps) {
  return (
    <div>
      <FieldLabel name={name} required={required} helpText={helpText} htmlFor={fieldName} complete={complete} />
      <input
        id={fieldName}
        type="number"
        value={value}
        onChange={(e) => onChange(fieldName, Number(e.target.value))}
        min={min}
        max={max}
        className="w-full border border-gray-600 bg-gray-800 text-gray-100 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
      />
    </div>
  );
}
