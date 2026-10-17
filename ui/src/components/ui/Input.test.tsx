/**
 * Input.test.tsx — the form primitives wire labels, hints and errors to their
 * controls, and the Toggle switch drives its hidden checkbox so a form sees
 * the same change a click on a native checkbox would make.
 */

import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  Checkbox,
  FormGroup,
  FormSection,
  Input,
  SearchInput,
  Select,
  Textarea,
  Toggle,
} from './Input';

afterEach(cleanup);

describe('labelled fields derive an id from the label when none is given', () => {
  it('Input', () => {
    render(<Input label="Frame size" hint="Bytes" />);
    expect(screen.getByLabelText('Frame size')).toHaveAttribute('id', 'frame-size');
    expect(screen.getByText('Bytes')).toBeInTheDocument();
  });

  it('Textarea', () => {
    render(<Textarea label="Run notes" />);
    expect(screen.getByLabelText('Run notes')).toHaveAttribute('id', 'run-notes');
  });

  it('Select', () => {
    render(<Select label="Test mode" options={[{ value: 'a', label: 'A' }]} />);
    expect(screen.getByLabelText('Test mode')).toHaveAttribute('id', 'test-mode');
  });

  it('Checkbox', () => {
    render(<Checkbox label="Reflect all" description="Every frame" />);
    expect(screen.getByLabelText('Reflect all')).toHaveAttribute('id', 'reflect-all');
    expect(screen.getByText('Every frame')).toBeInTheDocument();
  });
});

describe('an error replaces the hint', () => {
  it('Input', () => {
    render(<Input label="Rate" hint="Mbit/s" error="Too high" />);
    expect(screen.getByText('Too high')).toBeInTheDocument();
    expect(screen.queryByText('Mbit/s')).not.toBeInTheDocument();
  });

  it('Textarea', () => {
    render(<Textarea label="Notes" hint="Optional" error="Too long" />);
    expect(screen.getByText('Too long')).toBeInTheDocument();
    expect(screen.queryByText('Optional')).not.toBeInTheDocument();
  });

  it('Select', () => {
    render(<Select label="Mode" options={[]} hint="Pick one" error="Required" />);
    expect(screen.getByText('Required')).toBeInTheDocument();
    expect(screen.queryByText('Pick one')).not.toBeInTheDocument();
  });
});

it('Input renders both icons around the field', () => {
  render(<Input aria-label="Host" leftIcon={<span>L</span>} rightIcon={<span>R</span>} />);
  expect(screen.getByText('L')).toBeInTheDocument();
  expect(screen.getByText('R')).toBeInTheDocument();
});

describe('Select', () => {
  it('reports the chosen value, not the event', () => {
    const onChange = vi.fn();
    render(
      <Select
        label="Mode"
        placeholder="Choose"
        onChange={onChange}
        defaultValue=""
        options={[
          { value: 'tcp', label: 'TCP' },
          { value: 'udp', label: 'UDP', disabled: true },
        ]}
      />,
    );
    expect(screen.getByRole('option', { name: 'Choose' })).toBeDisabled();
    expect(screen.getByRole('option', { name: 'UDP' })).toBeDisabled();
    fireEvent.change(screen.getByLabelText('Mode'), { target: { value: 'tcp' } });
    expect(onChange).toHaveBeenCalledWith('tcp');
  });
});

describe('Toggle', () => {
  it('names the switch from its label and reflects the checked state', () => {
    render(<Toggle label="Automatic updates" description="Daily" checked onChange={() => {}} />);
    const toggle = screen.getByRole('switch', { name: 'Automatic updates' });
    expect(toggle).toHaveAttribute('aria-checked', 'true');
    expect(screen.getByText('Daily')).toBeInTheDocument();
  });

  it('clicking the switch fires the hidden checkbox change', () => {
    const onChange = vi.fn();
    render(<Toggle label="Automatic updates" checked={false} onChange={onChange} />);
    fireEvent.click(screen.getByRole('switch', { name: 'Automatic updates' }));
    expect(onChange).toHaveBeenCalledTimes(1);
  });
});

describe('SearchInput', () => {
  it('shows no clear button while empty', () => {
    render(<SearchInput label="Find" value="" onChange={() => {}} />);
    expect(screen.queryByRole('button', { name: 'Clear search' })).not.toBeInTheDocument();
    expect(screen.getByLabelText('Find')).toHaveAttribute('placeholder', 'Search...');
  });

  it('clears through onClear when one is given', () => {
    const onClear = vi.fn();
    const onChange = vi.fn();
    render(<SearchInput aria-label="Find" value="rfc" onChange={onChange} onClear={onClear} />);
    fireEvent.click(screen.getByRole('button', { name: 'Clear search' }));
    expect(onClear).toHaveBeenCalledTimes(1);
    expect(onChange).not.toHaveBeenCalled();
  });

  it('clears through onChange with an empty value otherwise', () => {
    const onChange = vi.fn();
    render(<SearchInput aria-label="Find" value="rfc" onChange={onChange} />);
    fireEvent.click(screen.getByRole('button', { name: 'Clear search' }));
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ target: { value: '' } }));
  });
});

it('FormSection renders its title, description and children inside a FormGroup', () => {
  render(
    <FormGroup>
      <FormSection title="Traffic" description="Stream settings">
        <span>child</span>
      </FormSection>
    </FormGroup>,
  );
  expect(screen.getByRole('heading', { name: 'Traffic' })).toBeInTheDocument();
  expect(screen.getByText('Stream settings')).toBeInTheDocument();
  expect(screen.getByText('child')).toBeInTheDocument();
});
