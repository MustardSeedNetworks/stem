/**
 * Input primitive stories (Wave 5 / #236).
 *
 * Covers label/hint/error decorations, leftIcon/rightIcon slots,
 * and the Input/Textarea/Select/Checkbox/Toggle/SearchInput sibling
 * primitives that ship from the same module.
 */
import type { Meta, StoryObj } from '@storybook/react-vite';
import { Lock, Mail } from 'lucide-react';
import { useState } from 'react';
import { Checkbox, FormGroup, FormSection, Input, SearchInput, Select, Textarea, Toggle } from './';

const meta: Meta<typeof Input> = {
  title: 'UI/Input',
  component: Input,
  parameters: { layout: 'centered' },
  argTypes: {
    label: { control: 'text' },
    placeholder: { control: 'text' },
    hint: { control: 'text' },
    error: { control: 'text' },
    disabled: { control: 'boolean' },
  },
};
export default meta;

type Story = StoryObj<typeof Input>;

export const Default: Story = {
  args: { label: 'Username', placeholder: 'admin' },
};

export const WithHint: Story = {
  args: { label: 'API token', placeholder: 'st_…', hint: 'Found in Settings → API' },
};

export const WithError: Story = {
  args: { label: 'Email', placeholder: 'you@example.com', error: 'Email is required' },
};

export const WithLeftIcon: Story = {
  args: { label: 'Email', placeholder: 'you@example.com', leftIcon: <Mail className="w-4 h-4" /> },
};

export const Password: Story = {
  args: {
    label: 'Password',
    type: 'password',
    placeholder: '••••••••',
    leftIcon: <Lock className="w-4 h-4" />,
  },
};

export const Disabled: Story = {
  args: { label: 'Locked field', value: 'cannot edit', disabled: true },
};

export const Search_: Story = {
  name: 'SearchInput',
  render: function searchRender() {
    const [q, setQ] = useState('');
    return (
      <SearchInput
        label="Find host"
        placeholder="Search hostname or IP…"
        value={q}
        onChange={(e) => setQ(e.target.value)}
        onClear={() => setQ('')}
      />
    );
  },
};

export const TextareaExample: Story = {
  name: 'Textarea',
  render: () => <Textarea label="Notes" placeholder="Optional details" rows={4} />,
};

export const SelectExample: Story = {
  name: 'Select',
  render: () => (
    <Select label="Mode" defaultValue="reflect">
      <option value="reflect">Reflector</option>
      <option value="benchmark">RFC 2544</option>
      <option value="y1564">Y.1564</option>
    </Select>
  ),
};

export const CheckboxExample: Story = {
  name: 'Checkbox',
  render: function checkboxRender() {
    const [checked, setChecked] = useState(true);
    return (
      <Checkbox
        label="Persist results to disk"
        checked={checked}
        onChange={(e) => setChecked(e.target.checked)}
      />
    );
  },
};

export const ToggleExample: Story = {
  name: 'Toggle',
  render: function toggleRender() {
    const [on, setOn] = useState(false);
    return <Toggle label="Strict mode" checked={on} onChange={(e) => setOn(e.target.checked)} />;
  },
};

export const FormComposition: Story = {
  name: 'Form composition',
  render: () => (
    <FormSection title="Run profile" description="Persisted at /etc/stem.conf">
      <FormGroup>
        <Input label="Name" placeholder="benchmark-01" />
        <Input label="Target" placeholder="10.0.0.2" />
        <Input label="Duration (s)" type="number" defaultValue={60} />
      </FormGroup>
    </FormSection>
  ),
};
