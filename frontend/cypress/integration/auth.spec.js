// auth.spec.js created with Cypress
//
// Start writing your Cypress tests below!
// If you're unfamiliar with how Cypress works,
// check out the link below and learn how to write your first test:
// https://on.cypress.io/writing-first-test




describe("Auth test", () => {

    before(() => {
        Cypress.Commands.add('login', () => {
            cy.session(["user"], () => {
                cy.visit('/login')
                cy.get('#username').should('be.visible').type('user')
                cy.get('#password').should('be.visible').type('user')
                cy.get('button[type="submit"]').should('be.visible').click()
                cy.get('main h5:contains("Latest documents")')
            })
        })
    })


    it('should login', () => {

        cy.login()

       /*
        cy.visit('/login')
        cy.get('#username').should('be.visible').type('user')
        cy.get('#password').should('be.visible').type('user')
        cy.get('button[type="submit"]').should('be.visible').click()

        cy.get('main h5:contains("Latest documents")')

        */
    })
})

