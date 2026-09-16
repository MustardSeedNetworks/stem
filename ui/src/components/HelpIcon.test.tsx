import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { HelpIcon } from './HelpIcon';

afterEach(cleanup);

describe('HelpIcon keyboard help', () => {
  it('describes the focused control and lets Escape dismiss without clicking', () => {
    const onClick = vi.fn();
    render(<HelpIcon tooltip="Frames sent each second" onClick={onClick} />);
    const button = screen.getByRole('button');
    expect(button).not.toHaveAttribute('title');
    fireEvent.focus(button);
    const tooltip = screen.getByRole('tooltip');
    expect(button).toHaveAttribute('aria-describedby', tooltip.id);
    expect(tooltip).toHaveTextContent('Frames sent each second');
    fireEvent.keyDown(button, { key: 'Escape' });
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
    expect(onClick).not.toHaveBeenCalled();
    fireEvent.click(button);
    expect(onClick).toHaveBeenCalledOnce();
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
  });

  it('dismisses hover help even when the trigger does not have focus', () => {
    render(<HelpIcon tooltip="Hover help" />);
    fireEvent.mouseEnter(screen.getByRole('button'));
    expect(screen.getByRole('tooltip')).toBeVisible();
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
  });

  it('does not toggle a label checkbox when help is clicked', () => {
    const onChange = vi.fn();
    render(
      <label>
        <input type="checkbox" onChange={onChange} />
        Rate
        <HelpIcon tooltip="Rate help" />
      </label>,
    );
    fireEvent.click(screen.getByRole('button'));
    expect(onChange).not.toHaveBeenCalled();
    expect(screen.getByRole('checkbox')).not.toBeChecked();
  });
});
