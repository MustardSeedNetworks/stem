import { act, cleanup, fireEvent, render, screen } from '@testing-library/react';
import type { ReactElement } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { Button } from './Button';
import { Tooltip } from './Tooltip';

afterEach(cleanup);

it('dismisses after an action focuses its opener for modal restoration', () => {
  render(
    <Tooltip text="Open help">
      <button type="button" onClick={(event) => event.currentTarget.focus()}>
        Help
      </button>
    </Tooltip>,
  );
  const button = screen.getByRole('button');
  fireEvent.mouseEnter(button);
  expect(screen.getByRole('tooltip')).toBeVisible();
  fireEvent.click(button);
  expect(button).toHaveFocus();
  expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
});

describe('an unavailable action stays reachable so its reason can be read', () => {
  const cases: { name: string; trigger: (onClick: () => void) => ReactElement }[] = [
    {
      name: 'native button',
      trigger: (onClick) => (
        <button type="button" disabled onClick={onClick}>
          Start
        </button>
      ),
    },
    {
      name: 'shared Button',
      trigger: (onClick) => (
        <Button disabled onClick={onClick}>
          Start
        </Button>
      ),
    },
  ];

  it.each(cases)('$name', ({ trigger }) => {
    const onClick = vi.fn();
    render(<Tooltip text="Not available on this platform">{trigger(onClick)}</Tooltip>);
    const button = screen.getByRole('button', { name: 'Start' });
    expect(button).toBeEnabled();
    expect(button).toHaveAttribute('aria-disabled', 'true');
    act(() => button.focus());
    expect(button).toHaveFocus();
    expect(button).toHaveAccessibleDescription('Not available on this platform');
    fireEvent.click(button);
    expect(onClick).not.toHaveBeenCalled();
  });
});
