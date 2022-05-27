// document.js created with Cypress
//
// Start writing your Cypress tests below!
// If you're unfamiliar with how Cypress works,
// check out the link below and learn how to write your first test:
// https://on.cypress.io/writing-first-test


describe("Document test", () => {

    before(() => {
        cy.login();
    })

    it('should upload pdf', () => {
        cy.visit('/documents')
        cy.get('span').contains('Create').click() ///.should('be.visible').click()
        cy.get('[data-testid="dropzone"]').attachFile('jpg-1.jpg', { subjectType: 'drag-n-drop' })
        cy.get('button').contains("Save").click()
    })
})
