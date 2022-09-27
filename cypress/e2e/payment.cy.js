describe('Payment', () => {
    describe('Pay for LPA', () => {
        it('adds a secure cookie before redirecting user to GOV UK Pay', () => {
            cy.getCookie('pay').should('not.exist')

            cy.visit('/testing-start?redirect=/about-payment');
            cy.injectAxe();

            cy.get('h1').should('contain', 'About payment');

            cy.checkA11y(null, { rules: { region: { enabled: false } } });

            cy.intercept('*/payment-confirmation', (req) => {
                cy.getCookie('pay').should('exist')
            })

            cy.contains('button', 'Continue to payment').click()
            cy.url().should('eq', `${Cypress.config('baseUrl')}/payment-confirmation`)
        })

        it('removes existing secure cookie on payment confirmation page', () => {
            cy.setCookie('pay', 'abc123')
            cy.visit('/testing-start?redirect=/payment-confirmation&paymentComplete=1');
            cy.injectAxe();

            cy.get('h1').should('contain', 'Payment received');
            cy.checkA11y(null, { rules: { region: { enabled: false } } });
            cy.getCookie('pay').should('not.exist')

            cy.contains('a', 'Continue').click()
            // Will lead to identity journey once we have an initial page
            cy.url().should('eq', `${Cypress.config('baseUrl')}/task-list`)

        })
    })
})
