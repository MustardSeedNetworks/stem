/**
 * useFocusTrap had no coverage, which is the wrong state for the code that
 * decides whether a keyboard user can get out of a drawer: a regression here
 * is invisible to a mouse and total to a keyboard.
 *
 * These pin the wrapping behaviour specifically, because the guard that finds
 * the trap's two ends was restructured for noUncheckedIndexedAccess —
 * `focusableElements.length === 0` became a check on the ends themselves. The
 * conditions are equivalent, and equivalence is worth proving rather than
 * asserting.
 */
import { fireEvent } from '@testing-library/dom';
import { render, screen } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { useFocusTrap } from './useFocusTrap';

/**
 * The hook skips candidates whose `offsetParent` is null, so it ignores
 * elements that are not laid out. jsdom implements no layout and reports null
 * for everything, so without this the trap silently finds nothing focusable
 * and every assertion below would pass for the wrong reason — which is how a
 * hook like this ends up with no tests in the first place.
 *
 * The visibility filter itself therefore cannot be tested here; it needs a
 * real browser, and belongs in the Playwright suite.
 */
let realOffsetParent: PropertyDescriptor | undefined;

beforeAll(() => {
  realOffsetParent = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetParent');
  Object.defineProperty(HTMLElement.prototype, 'offsetParent', {
    configurable: true,
    get(): Element | null {
      return document.body;
    },
  });
});

afterAll(() => {
  if (realOffsetParent) {
    Object.defineProperty(HTMLElement.prototype, 'offsetParent', realOffsetParent);
  }
});

function Trapped({ empty = false }: { empty?: boolean }) {
  const ref = useFocusTrap<HTMLDivElement>({ isActive: true, autoFocus: false });
  return (
    <div ref={ref} data-testid="trap">
      {empty ? null : (
        <>
          <button type="button">first</button>
          <button type="button">middle</button>
          <button type="button">last</button>
        </>
      )}
    </div>
  );
}

describe('useFocusTrap', () => {
  it('wraps Tab from the last element back to the first', () => {
    render(<Trapped />);
    screen.getByRole('button', { name: 'last' }).focus();

    fireEvent.keyDown(document, { key: 'Tab' });

    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'first' }));
  });

  it('wraps Shift+Tab from the first element to the last', () => {
    render(<Trapped />);
    screen.getByRole('button', { name: 'first' }).focus();

    fireEvent.keyDown(document, { key: 'Tab', shiftKey: true });

    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'last' }));
  });

  it('leaves Tab alone in the middle of the list', () => {
    render(<Trapped />);
    const middle = screen.getByRole('button', { name: 'middle' });
    middle.focus();

    fireEvent.keyDown(document, { key: 'Tab' });

    /* The browser moves focus itself; the trap only intervenes at the ends. */
    expect(document.activeElement).toBe(middle);
  });

  it('does nothing at all when the container has nothing focusable', () => {
    render(
      <>
        <button type="button">outside</button>
        <Trapped empty={true} />
      </>,
    );
    const outside = screen.getByRole('button', { name: 'outside' });
    outside.focus();

    /* The empty case is the one the restructured guard covers: with no ends
       to trap between, the handler returns rather than reaching for one. */
    expect(() => fireEvent.keyDown(document, { key: 'Tab' })).not.toThrow();
    expect(document.activeElement).toBe(outside);
  });
});
