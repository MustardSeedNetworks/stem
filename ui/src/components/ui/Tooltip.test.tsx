import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, expect, it } from 'vitest';
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
