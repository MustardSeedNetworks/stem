import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, userEvent, within } from 'storybook/test';
import { Tooltip } from './Tooltip';

const meta: Meta<typeof Tooltip> = {
  title: 'UI/Tooltip',
  component: Tooltip,
};
export default meta;
type Story = StoryObj<typeof Tooltip>;

export const UnavailableAction: Story = {
  render: function UnavailableAction() {
    const [disabled, setDisabled] = useState(true);
    const [clicks, setClicks] = useState(0);
    const [submits, setSubmits] = useState(0);
    return (
      <form
        onSubmit={(event) => {
          event.preventDefault();
          setSubmits((value) => value + 1);
        }}
      >
        <Tooltip text="Choose an interface before starting">
          <button type="submit" disabled={disabled} onClick={() => setClicks((value) => value + 1)}>
            Start test
          </button>
        </Tooltip>
        <button type="button" onClick={() => setDisabled(false)}>
          Choose interface
        </button>
        <output data-testid="actions">
          {clicks} clicks, {submits} submissions
        </output>
      </form>
    );
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const action = canvas.getByRole('button', { name: 'Start test' });
    await expect(action).toHaveAttribute('aria-disabled', 'true');
    action.focus();
    await expect(action).toHaveFocus();
    await expect(action).toHaveAccessibleDescription('Choose an interface before starting');
    await userEvent.keyboard('{Enter}');
    await userEvent.keyboard(' ');
    await userEvent.click(action);
    await expect(canvas.getByTestId('actions')).toHaveTextContent('0 clicks, 0 submissions');
    await userEvent.click(canvas.getByRole('button', { name: 'Choose interface' }));
    await userEvent.click(action);
    await expect(canvas.getByTestId('actions')).toHaveTextContent('1 clicks, 1 submissions');
  },
};
