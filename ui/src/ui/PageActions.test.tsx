import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import '../i18n';
import { PageActionsSlotContext, RefreshAction } from './PageActions';

afterEach(cleanup);

describe('RefreshAction', () => {
  it('appears in the header slot and reloads what the page loaded', () => {
    const slot = document.createElement('div');
    slot.setAttribute('data-testid', 'slot');
    document.body.append(slot);
    const onRefresh = vi.fn();

    render(
      <PageActionsSlotContext.Provider value={slot}>
        <RefreshAction onRefresh={onRefresh} />
      </PageActionsSlotContext.Provider>,
    );

    const button = screen.getByTestId('page-refresh');
    expect(slot).toContainElement(button);
    expect(button).toHaveTextContent('Refresh');
    fireEvent.click(button);
    expect(onRefresh).toHaveBeenCalledOnce();
    slot.remove();
  });

  it('renders nothing before the header exists', () => {
    render(<RefreshAction onRefresh={() => undefined} />);

    expect(screen.queryByTestId('page-refresh')).toBeNull();
  });
});
